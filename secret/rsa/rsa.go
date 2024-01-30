// Copyright 2024 The bee Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	// DefaultPrivateKey private key
	// openssl genrsa -out keys.pem 1024
	DefaultPrivateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDkv765UFFGyCJJWlPT+2ZeOxsw9ZPmQGniwq0/blxwvRGUlfzo
C4x8t2QY0crm1sZP6P93HXjg09bOsHbQT04itoyCY977/dd0nBIgZ/qVkehymMoA
tclkoAqNzTXYNzp5DZ6kRMtlmyX7EgndRPQ3Mm9cNd5q+paOUCAwEiddGwIDAQAB
AoGAKo7mDBI+XN3lSyJsEDdor0Vt5Kj78E2xpYe0teVxe2QhvjQ9jmp/o8B29gsq
JdJ1qO5fgSjRkXq4L1IzeMQYdBDMxqG9CGSufWll0LtSmNAIBm6AKNO4dA74OzpC
UO7nzX+djGb02ZG5tKRQ4mMuLW/2PwoSepfWccwAzc7np1ECQQD5/v9bUqtaz7Cw
eyMBLpNNp9sJNS0RTfz9EwpIyynOV8CvOJzvRfHGb2xtHGqSIFD2ptb5zysBe9/v
D46HTAIpAkEA6j4eQ7Ms2GH7TyV8EL/0WrM39OTa85Z5DdmBvkpSrM/mTGd/e0mF
E8c9tJ8JFswTdIqKj5HEEjF4GJNKesO1owJBAIDMnPmLBRe7a3fxaR6BxYi704DR
8c85k/87IRBSA874rSBZlZk9OwyWeZFZk5qHpc7+NEHuN2UDUmNTa4ZPZckCQQCn
UqQPvAeGscbwbFhJJrUHrQmFl4yHf68NI5e4NCMGaqOZZDz99jBnRmVfhlLZxAEJ
uITttTQXwtqEw4HqW659AkBNZIelmCJL9zFV1VcXOgzuO870a2zm/hodxs9ocndk
2BENmtxu78U6IHLB3GzuWUiBXP1RLms/4Vd3Q4MxUyyb
-----END RSA PRIVATE KEY-----
`)

	// DefaultPublicKey public key
	// openssl rsa -in keys.pem -pubout -out public.pem
	DefaultPublicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDkv765UFFGyCJJWlPT+2ZeOxsw
9ZPmQGniwq0/blxwvRGUlfzoC4x8t2QY0crm1sZP6P93HXjg09bOsHbQT04itoyC
Y977/dd0nBIgZ/qVkehymMoAtclkoAqNzTXYNzp5DZ6kRMtlmyX7EgndRPQ3Mm9c
Nd5q+paOUCAwEiddGwIDAQAB
-----END PUBLIC KEY-----
`)
)

func Encode(text []byte) ([]byte, error) {
	block, _ := pem.Decode(DefaultPublicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, text)
}

func Decode(text []byte) ([]byte, error) {
	block, _ := pem.Decode(DefaultPrivateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, text)
}
