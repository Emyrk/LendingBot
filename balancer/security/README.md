# Generate Cert Pair

```
openssl req -newkey rsa:4096 -nodes -sha512 -x509 -days 36500 -nodes -out ./cert.pem -keyout ./key.pem
```