# simpleredirect
A simple URL shrinking &amp; redirecting service, written in Golang with MySQL database.
**Current Version: v1.00**

**First installation guide**

-> Download directory to '$GOPATH/src'

-> Download all external dependencies

-> Create link of folder "$GOPATH/src/simpleredirect/simpleredirecttmp" in directory '$GOPATH/bin'

-> Rename link "simpleredirecttmp"

**->** Create necessary databases. I will add the SQL files in upcoming commits.

-> *Run in terminal:*

$ go install simpleredirect

**Execute created executable file**

-> *Run in terminal:*

$ cd $GOPATH/bin

$ ./simpleredirect

**Version Releases:**
1. 1.00 : 19th July, 2016

First release. Fully functional application, provisioned with user authentication, session management.
