package main

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
)

func testStores(t *testing.T, s Store) {
	usersM := []User{
		User{
			UserFull: UserFull{
				UserMeta: UserMeta{
					ID:   "user1",
					Name: "User 1",
				},
				PublicKeys: map[string][]byte{
					"pubkey1": []byte(`pubkey1`),
					"pubkey2": []byte(`pubkey2`),
				},
			},
			Password:              []byte("user1"),
			RequiresPasswordReset: true,
			PrivateKeys: map[string][]byte{
				"prikey1": []byte(`prikey1`),
				"prikey2": []byte(`prikey2`),
			},
		},
		User{
			UserFull: UserFull{
				UserMeta: UserMeta{
					ID:   "user2",
					Name: "User 2",
				},
				PublicKeys: map[string][]byte{
					"pubkey3": []byte(`pubkey3`),
					"pubkey4": []byte(`pubkey4`),
				},
			},
			Password:              []byte("user2"),
			RequiresPasswordReset: false,
			PrivateKeys: map[string][]byte{
				"prikey3": []byte(`prikey3`),
				"prikey4": []byte(`prikey4`),
			},
		},
	}

	for _, u := range usersM {
		var uu User
		uu.UserMeta = u.UserMeta
		uu.Password = u.Password
		if err := s.PostUser(u); err != nil {
			t.Fatalf("Got unexpected error when creating user %q: %v", u.ID, err)
		}
		for k, v := range u.PublicKeys {
			if err := s.AddPublicKey(u.ID, k, v); err != nil {
				t.Fatalf("Got unexpected error when adding public key %q to user %q: %v", k, u.ID, err)
			}
		}
		for k, v := range u.PrivateKeys {
			if err := s.AddPrivateKey(u.ID, k, v); err != nil {
				t.Fatalf("Got unexpected error when adding private key %q to user %q: %v", k, u.ID, err)
			}
		}
	}

	if err := s.AddExternalPublicKey("pubkeyExt", []byte("pubkeyExt")); err != nil {
		t.Fatalf("Got unexpected error when adding external public key: %v", err)
	}

	if users, err := s.ListUsers(); err != nil {
		t.Fatal("Got unexpected error when listing users:", err)
	} else if len(users) != 2 {
		t.Fatalf("Found %d users, expected %d", len(users), 2)
	}

	for _, u := range usersM {
		if uu, err := s.GetUser(u.ID); err != nil {
			t.Fatalf("Got unexpected error when finding user %q: %v", u.ID, err)
		} else if uu.ID != u.ID || uu.Name != u.Name || !bytes.Equal(uu.Password, u.Password) || uu.RequiresPasswordReset != u.RequiresPasswordReset {
			t.Fatalf("Didn't get back correct user.")
		}
	}

	for u, keys := range map[string][]string{
		"user1": []string{"pubkey1", "pubkey2"},
		"user2": []string{"pubkey3", "pubkey4"},
	} {
		if k, err := s.GetPublicKeyIDs(u); err != nil {
			t.Fatalf("Got unexpected error when finding public keys for user %q: %v", u, err)
		} else {
			sort.Strings(k)
			sort.Strings(keys)
			if !reflect.DeepEqual(k, keys) {
				t.Fatalf("Didn't get expected list of public keys for user %q: %v != %v", u, k, keys)
			}
		}

		if m, err := s.GetPublicKeys(u); err != nil {
			t.Fatalf("Got unexpected error when finding public keys for user %q: %v", u, err)
		} else if len(m) != 2 {
			t.Fatalf("Got unexpected number of public keys for user %q: %d != %d", u, len(m), 2)
		} else {
			for _, k := range keys {
				if !bytes.Equal(m[k], []byte(k)) {
					t.Fatalf("Didn't get expected public key for user %q: %q != %q", u, m[k], k)
				}
			}
		}
	}

	for u, keys := range map[string][]string{
		"user1": []string{"prikey1", "prikey2"},
		"user2": []string{"prikey3", "prikey4"},
	} {
		if k, err := s.GetPrivateKeyIDs(u); err != nil {
			t.Fatalf("Got unexpected error when finding private keys for user %q: %v", u, err)
		} else {
			sort.Strings(k)
			sort.Strings(keys)
			if !reflect.DeepEqual(k, keys) {
			}
		}

		if m, err := s.GetPrivateKeys(u); err != nil {
			t.Fatalf("Got unexpected error when finding private keys for user %q: %v", u, err)
		} else if len(m) != 2 {
			t.Fatalf("Got unexpected number of private keys for user %q: %d != %d", u, len(m), 2)
		} else {
			for _, k := range keys {
				if !bytes.Equal(m[k], []byte(k)) {
					t.Fatalf("Didn't get expected private key for user %q: %q != %q", u, m[k], k)
				}
			}
		}
	}

	for k, u := range map[string]string{
		"pubkey1":   "user1",
		"pubkey2":   "user1",
		"pubkey3":   "user2",
		"pubkey4":   "user2",
		"pubkeyExt": "",
	} {
		if uu, err := s.GetUserForPublicKey(k); err != nil {
			t.Fatalf("Got unexpected error when finding user for public key %q: %v", k, err)
		} else if uu != u {
			t.Fatalf("Got wrong user when finding user for public key: %v != %v", uu, u)
		}
	}

	if err := s.RemovePrivateKey("user2", "prikey3"); err != nil {
		t.Fatal("Got unexpected error when removing private key:", err)
	} else if keys, err := s.GetPrivateKeyIDs("user2"); err != nil {
		t.Fatal("Got unexpected error when getting private keys:", err)
	} else if len(keys) != 1 || keys[0] != "prikey4" {
		t.Fatalf("Got unexpected list of private keys after deleting: %v != %v", keys, []string{"prikey4"})
	}

	if err := s.RemovePublicKey("user1", "pubkey2"); err != nil {
		t.Fatal("Got unexpected error when removing public key:", err)
	} else if keys, err := s.GetPublicKeyIDs("user1"); err != nil {
		t.Fatal("Got unexpected error when getting public keys:", err)
	} else if len(keys) != 1 || keys[0] != "pubkey1" {
		t.Fatalf("Got unexpected list of public keys after deleting: %v != %v", keys, []string{"pubkey1"})
	} else if u, err := s.GetUserForPublicKey("pubkey2"); err != nil {
		t.Fatal("Got unexpected error when getting user for deleted public key:", err)
	} else if u != "" {
		t.Fatalf("Got unexpected user for deleted public keys: %q != %q", u, "")
	}

	if err := s.PutUser(User{
		UserFull: UserFull{
			UserMeta: UserMeta{
				ID:   "user1",
				Name: "foobar",
			},
		},
		Password:              []byte("user1"),
		RequiresPasswordReset: false,
	}); err != nil {
		t.Fatal("Got unexpected error when modifying user:", err)
	} else if u, err := s.GetUser("user1"); err != nil {
		t.Fatal("Got unexpected error when getting user:", err)
	} else if u.Name != "foobar" {
		t.Fatalf("Name didn't update when modifying user: %q != %q", u.Name, "foobar")
	} else if !bytes.Equal(u.Password, []byte("user1")) {
		t.Fatalf("Password didn't update when modifying user: %q != %q", u.Password, "user1")
	} else if u.RequiresPasswordReset != false {
		t.Fatalf("RequiresPasswordReset didn't update when modifying user: %v != %v", u.RequiresPasswordReset, false)
	}

	if err := s.DeleteUser("user1"); err != nil {
		t.Fatal("Got unexpected error when removing user:", err)
	} else if _, err := s.GetUser("user1"); err == nil {
		t.Fatal("Could still get user after removing")
	} else if u, err := s.GetUserForPublicKey("pubkey1"); err != nil {
		t.Fatal("Got unexpected error when getting user for deleted public key:", err)
	} else if u != "" {
		t.Fatalf("Got unexpected user for public keys owned by a deleted user: %q != %q", u, "")
	} else if keys, err := s.GetPrivateKeyIDs("user1"); err != nil {
		t.Fatal("Got unexpected error when getting user for deleted public key:", err)
	} else if len(keys) != 0 {
		t.Fatal("Got unexpected private keys for a deleted user")
	}
}
