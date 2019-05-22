# gophercoin
a simple blockchain and cryptocurrency implementation


# todo

### Use `http.ListenAndServeTLS()` to serve the API
#### Generate private key (.key)
`# Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out api.key 2048

# Key considerations for algorithm "ECDSA" ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl ecparam -genkey -name secp384r1 -out api.key`
#### Generation of self-signed(x509) public key (PEM-encodings `.pem`|`.crt`) based on the private (`.key`)
`
openssl req -new -x509 -sha256 -key api.key -out api.crt -days 3650
`
### 