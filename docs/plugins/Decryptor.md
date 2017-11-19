A decryptor allows you to use encrypted versions of the passwords for the MySQL
database and Spinnaker p12 certificate (see [configuration file format](Configuration-file-format)).
Chaos Monkey will invoke the decryptor to decrypt the passwords before using
them.

Chaos Monkey does not ship with any decryptor implementations. If you wish to
use this functionality, you will need to implement your own.


If you wish to store your passwords encrypted and use a decryption system at
runtime, you need to:

1. Give your decryptor a name (e.g., "gpg")
1. Code up a type in Go that implements the [Decryptor](https://godoc.org/github.com/Netflix/chaosmonkey/#Decryptor) interface.
1. Modify [decryptor.go](https://github.com/Netflix/chaosmonkey/blob/master/decryptor/decryptor.go) so that it recognizes your decryptor.
1. Edit your [config file](Configuration-file-format) to specify your decryptor.


