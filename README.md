# GoPasswordManager

A password manager written in Go and Angular.js. Passwords are encrypted in-browser with user-supplied GPG keys. The application has been designed to be compatible with the [pass](https://www.passwordstore.org/) program.
Users can clone the git repository `password-store.git` located in the application root after the application has launched an interact with it before pushing changes.
The application has default credentials
 - Username: tolar2
 - Password: tolar2

The default GPG key has the same password.

To run the application, you need to install [Golang](https://golang.org/), [Node.js](https://nodejs.org/en/), and [a GCC compiler if you're on Windows](http://tdm-gcc.tdragon.net/). Clone the repo, open a command prompt and navigate to the folder, then run

    $ cd path/to/repo/GoPasswordManager
    $ cd app
    $ npm install
    $ cd ..
    $ go get .
    $ go build .
    $ ./GoProgramManager or .\GoProgramManager.exe
    
And navigate to http://localhost:8080/
