# MongoDB source - Install MongoDB Community Edition on macOS
#https://docs.mongodb.com/manual/tutorial/install-mongodb-on-os-x/
#Install MongoDB on Mac
brew tap mongodb/brew
brew install mongodb-community@4.2

#Start MongoDB on Mac using below
brew services start mongodb-community@4.2
# OR 
mongod --config /usr/local/etc/mongod.conf --fork

#Good one 
#ca.key: Certificate Authority private key file (this shouldn't be shared in real-life)
#ca.cert: Certificate Authority trust certificate (this should be shared with users in real-life)
#server.key: Server private key, password protected (this shouldn't be shared)
#server.csr: Server certificate signing request (this should be shared with the CA owner)
#server.cert: Server certificate signed by the CA (this would be sent back by the CA owner) - keep on server
#server.pem: Conversion of server.key into format gRPC likes (this shouldn't be shared)

#Private files: ca.key, server.key, server.pem, server.crt
#'Share' files: ca.crt(needed by the client), server.csr(needed by the CA)

#Changes these CN's to atch your hosts in your environment as needed
SERVER_CN=localhost

#Step 1: Generate Certificate Authority + Trust Certificate (ca.crt)
openssl genrsa -passout pass:1111 -des3 -out ca.key 4096
openssl req -passin pass:1111 -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=${SERVER_CN}"

#Step 2: Generate the Server Private key (server.key)
openssl genrsa -passout pass:1111 -des3 -out server.key 4096

#Step 3: Get a certificate siging request from the CA (server.csr)
openssl req -passin pass:1111 -new -key server.key -out server.csr -subj "/CN=${SERVER_CN}"

#Step 4: Sign the certificate with the CA we created (it's called self signing) - server.crt
openssl x509 -req -passin pass:1111 -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt 

#Step 5: Convert the server certificate to .pem format (server.pem) - usable by gRPC
openssl pkcs8 -topk8 -nocrypt -passin pass:1111 -in server.key -out server.pem