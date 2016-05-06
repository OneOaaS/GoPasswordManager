package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/speedata/gogit"
)

const (
	recipientFile   = ".gpg-id"
	gitVerboseDeubg = false
)

type GitPass struct {
	repoRoot string
	branch   string
	debug    bool

	repo *gogit.Repository
}

type GitError struct {
	Err    error
	Stderr []byte
}

func (e GitError) Error() string {
	return fmt.Sprintf("%s\n%s", e.Err, e.Stderr)
}

func (g *GitPass) gitDebug(args ...string) {
	if g.debug {
		log.Print("git", args)
	}
}

func (g *GitPass) gitHelper(args ...string) *exec.Cmd {
	g.gitDebug(args...)
	cmd := exec.Command("git", args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Dir = g.repoRoot
	return cmd
}

func (g *GitPass) git(args ...string) error {
	var stderr bytes.Buffer
	cmd := g.gitHelper(args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return GitError{err, stderr.Bytes()}
	}
	return nil
}

func (g *GitPass) gitO(args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := g.gitHelper(args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return bytes.TrimSpace(stdout.Bytes()), GitError{err, stderr.Bytes()}
	}
	return bytes.TrimSpace(stdout.Bytes()), nil
}

func NewGitPass(root, branch string, debug bool) (*GitPass, error) {
	if branch == "" {
		branch = "master"
	}
	g := &GitPass{
		repoRoot: root,
		branch:   branch,
		debug:    debug,
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		// try to automatically create it
		if err := os.MkdirAll(root, 0700); err != nil {
			return nil, err
		} else if err := g.git("init", "--bare", "."); err != nil {
			return nil, err
		} else if b, err := g.gitO("mktree"); err != nil {
			return nil, err
		} else if b, err := g.gitO("commit-tree", string(b), "-m", "Initial Commit"); err != nil {
			return nil, err
		} else if err := g.git("update-ref", "refs/heads/"+g.branch, string(b)); err != nil {
			return nil, err
		}
	}

	// we should now have a git repo
	if repo, err := gogit.OpenRepository(root); err != nil {
		return nil, err
	} else if _, err := repo.LookupReference("refs/heads/" + g.branch); err != nil {
		return nil, err
	} else {
		g.repo = repo
	}

	return g, nil
}

type gitPassTx struct {
	g      *GitPass
	repo   *gogit.Repository
	branch string
	commit *gogit.Commit
	root   *gogit.Tree
}

type gitPassTxW struct {
	*gitPassTx

	// a nil entry indicates file deletion; a non-nil slice of zero length
	// indicates truncation
	changedPasswords map[string][]byte
	// a slice of zero length indicates removal
	changedRecipients map[string][]string
}

func (g *GitPass) Begin() (PassTx, error) {
	tx := &gitPassTx{
		g:      g,
		repo:   g.repo,
		branch: g.branch,
	}

	if ref, err := g.repo.LookupReference("refs/heads/" + g.branch); err != nil {
		return nil, err
	} else if c, err := g.repo.LookupCommit(ref.Oid); err != nil {
		return nil, err
	} else {
		tx.commit = c
		tx.root = c.Tree
		return tx, nil
	}
}

func (g *GitPass) BeginW() (PassTxW, error) {
	if txr, err := g.Begin(); err != nil {
		return nil, err
	} else {
		return &gitPassTxW{
			gitPassTx:         txr.(*gitPassTx),
			changedPasswords:  make(map[string][]byte),
			changedRecipients: make(map[string][]string),
		}, nil
	}
}

func (tx *gitPassTx) clean(p string) string {
	p = path.Clean(p)
	if p == "." {
		return ""
	} else {
		return strings.TrimPrefix(p, "/")
	}
}

func (tx *gitPassTx) getFile(p string) (*gogit.TreeEntry, error) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("getFile(%q)", p)
		defer log.Printf("getFile(%q) done", p)
	}
	if p == "" {
		return &gogit.TreeEntry{
			Type:     gogit.ObjectTree,
			Id:       tx.root.Oid,
			Filemode: gogit.FileModeTree,
			Name:     "",
		}, nil
	} else {
		dir, file := path.Split(p)
		dir = tx.clean(dir)
		if parent, err := tx.getFile(dir); err != nil {
			return nil, err
		} else if parent.Type != gogit.ObjectTree {
			return nil, os.ErrNotExist
		} else if t, err := tx.repo.LookupTree(parent.Id); err != nil {
			return nil, err
		} else if te := t.EntryByName(file); te == nil {
			return nil, os.ErrNotExist
		} else {
			return te, nil
		}
	}
}

func (tx *gitPassTx) Type(p string) (exists bool, file bool) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("Type(%q)", p)
		defer log.Printf("Type(%q) done", p)
	}

	p = tx.clean(p)

	if te, err := tx.getFile(p); err != nil {
		return false, false
	} else {
		return true, te.Type != gogit.ObjectTree
	}
}

func (tx *gitPassTx) List(p string) ([]PassDirent, error) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("List(%q)", p)
		defer log.Printf("List(%q) done", p)
	}

	p = tx.clean(p)

	if te, err := tx.getFile(p); err != nil {
		return nil, err
	} else if te.Type != gogit.ObjectTree {
		return nil, os.ErrInvalid
	} else if t, err := tx.repo.LookupTree(te.Id); err != nil {
		return nil, err
	} else {
		ret := make([]PassDirent, 0, len(t.TreeEntries))
		for _, te := range t.TreeEntries {
			// ignore dot files
			if strings.HasPrefix(te.Name, ".") {
				continue
			}
			ret = append(ret, PassDirent{
				Name: te.Name,
				File: te.Type == gogit.ObjectBlob,
			})
		}
		return ret, nil
	}
}

func (tx *gitPassTx) get(p string) ([]byte, error) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("get(%q)", p)
		defer log.Printf("get(%q) done", p)
	}

	if te, err := tx.getFile(p); err != nil {
		return nil, err
	} else if te.Type != gogit.ObjectBlob {
		return nil, os.ErrInvalid
	} else if b, err := tx.repo.LookupBlob(te.Id); err != nil {
		return nil, err
	} else {
		return b.Contents(), nil
	}
}

func (tx *gitPassTx) Get(p string) ([]byte, error) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("Get(%q)", p)
		defer log.Printf("Get(%q) done", p)
	}

	p = tx.clean(p)
	return tx.get(p)
}

func (tx *gitPassTx) recipients(p string, override map[string][]string) ([]string, error) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("recipients(%q)", p)
		defer log.Printf("recipients(%q) done", p)
	}

	// TODO: make this faster (each getFile starts from the root...)
	r := path.Join(p, recipientFile)
	if s := override[r]; len(s) > 0 {
		return s, nil
	} else if b, err := tx.get(r); err == nil {
		return strings.Split(strings.TrimSpace(string(b)), "\n"), nil
	} else if p != "" {
		dir, _ := path.Split(p)
		dir = strings.TrimSuffix(dir, "/")
		return tx.recipients(dir, override)
	} else {
		return nil, nil
	}
}

func (tx *gitPassTx) Recipients(p string) ([]string, error) {
	if gitVerboseDeubg && tx.g.debug {
		log.Printf("Recipients(%q)", p)
		defer log.Printf("Recipients(%q) done", p)
	}

	p = tx.clean(p)
	return tx.recipients(p, nil)
}

func (tx *gitPassTx) getAffectedFiles(p string, override map[string][]string) ([]string, error) {
	var ret []string
	err := tx.Walk(p, func(d PassDirent) error {
		if d.Name == p {
			return nil
		}
		if !d.File {
			// directory; check if it has a .gpg-id
			r := path.Join(d.Name, recipientFile)
			if s := override[r]; len(s) > 0 {
				return filepath.SkipDir
			} else if te, err := tx.getFile(r); err == nil && te.Type == gogit.ObjectBlob {
				return filepath.SkipDir
			} else {
				return nil
			}
		} else {
			ret = append(ret, d.Name)
		}
		return nil
	})
	return ret, err
}

func (tx *gitPassTx) GetAffectedFiles(p string) ([]string, error) {
	p = tx.clean(p)
	return tx.getAffectedFiles(p, nil)
}

func (tx *gitPassTx) Walk(p string, fn PassWalkFn) error {
	var q []*gogit.TreeEntry
	var qp []string
	p = tx.clean(p)

	if te, err := tx.getFile(p); err != nil {
		return err
	} else {
		q = append(q, te)
		qp = append(qp, p)
	}

	for len(q) > 0 {
		var te *gogit.TreeEntry
		var p string
		te, q = q[0], q[1:]
		p, qp = qp[0], qp[1:]
		isFile := te.Type == gogit.ObjectBlob
		if err := fn(PassDirent{
			Name: p,
			File: isFile,
		}); err == filepath.SkipDir {
			continue
		} else if err != nil {
			return err
		}

		if te.Type == gogit.ObjectTree {
			if t, err := tx.repo.LookupTree(te.Id); err != nil {
				return err
			} else {
				for _, te := range t.TreeEntries {
					if strings.HasPrefix(te.Name, ".") {
						continue
					}
					q = append(q, te)
					qp = append(qp, path.Join(p, te.Name))
				}
			}
		}
	}
	return nil
}

func (tx *gitPassTxW) SetRecipients(p string, recipients []string) {
	p = tx.clean(p)
	tx.changedRecipients[path.Join(p, recipientFile)] = recipients
	return
}

func (tx *gitPassTxW) Put(p string, contents []byte) {
	p = tx.clean(p)
	if len(contents) == 0 {
		contents = []byte{} // nil slices are distinct from zero-length
	}
	tx.changedPasswords[p] = contents
	return
}

func (tx *gitPassTxW) Delete(p string) {
	p = tx.clean(p)
	tx.changedPasswords[p] = nil
	return
}

func (tx *gitPassTxW) verify() error {
	// TODO(tolar2): actually verify...
	return nil
}

func (tx *gitPassTxW) Commit(message string) error {
	if err := tx.verify(); err != nil {
		return err
	}

	if message == "" {
		message = "Update passwords"
	}

	cmd := tx.g.gitHelper("fast-import")
	pw, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var w io.Writer = pw
	// var w io.Writer = io.MultiWriter(pw, os.Stdout)
	// var w io.Writer = os.Stdout

	fmt.Fprintf(w, "commit refs/heads/%s\n", tx.branch)
	now := time.Now()
	fmt.Fprintf(w, "committer Pass <pass@localhost> %d %s\n", now.Unix(), now.Format("-0700"))
	fmt.Fprintf(w, "data %d\n%s\n", len(message), message)
	fmt.Fprintf(w, "from %s\n", tx.commit.Oid)

	for n, r := range tx.changedRecipients {
		b := strings.Join(r, "\n")
		fmt.Fprintf(w, "M 644 inline %s\n", n)
		fmt.Fprintf(w, "data %d\n%s\n", len(b), b)
	}

	for n, pass := range tx.changedPasswords {
		if pass == nil {
			fmt.Fprintf(w, "D %s\n", n)
		} else {
			fmt.Fprintf(w, "M 644 inline %s\n", n)
			fmt.Fprintf(w, "data %d\n", len(pass))
			w.Write(pass)
			fmt.Fprint(w, "\n")
		}
	}
	fmt.Fprint(w, "done\n")

	pw.Close()

	return cmd.Wait()
}
