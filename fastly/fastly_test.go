package fastly

import (
	"strings"
	"testing"
)

// pgpPublicKey returns a PEM encoded PGP public key suitable for testing.
func pgpPublicKey() string {
	return `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBF7tJtMBCAC0gUVRRO8x5ViZbjHljHj91MNhspi43nLauAFgyD+kFWwZUtGz
vbesuUKpnManIXvllbzxvy5GgOvKZx4B8nKlvtTNeHmCBbPsHijrNvFvM7uhj2LF
ogY+IYT8RgtpiBh/Zuxgf7UzUL2tAX4fNELoU/mGJnW/7Qf0UpEujgV8+57OnlgA
YJcZvwCdN+Jv1xNaIzkue1SAsrbMwY9HMhy2NKEEzssMmIW+lictfNa+oMmEpxcn
GiDYgbLbPBuPIz13YbSN8n4e1qb9wlzTJQI3BapOx60IEFniwiAgToYkN1stU0BU
OrmtKvwqu3RPwPysWyobwu1h0DM3KducDSe9ABEBAAG0G0Zvb0JhciA8Zm9vYmFy
QGV4YW1wbGUuY29tPokBTgQTAQgAOAIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIX
gBYhBEkaBU4NHL/XNxtWU8Cg7XNtEc1bBQJe7SbzAAoJEMCg7XNtEc1bOqMH/j0Y
O41qfUFqkOVPWHXzPK4VxaGtTY4A3xkUtuEUgZojcpWLvP5ujqJ+VUjBo0AIRuBr
DnngWIZu7mfFMzQ5gq+8hXsh6PfT4V8ijJ9+2KadPYLeDlVNxzwQRugE0NVls/5v
Dj91OeZ1KRzTPzzgr2NI2nlyb6Li2I23zaz0Tj5kd6s1xZ/fuf/o6gs5YiNkIATQ
EAL3I+Wj6O97NsgYKYBy85cf0hMtF4hevih6UGdpxpSfM6+nOltDprAATEDkNsjY
O4YSAj1FwygpFE01x4RS0pDE4Z8VtnF9v29Y4rqWVamJrii1mQ6HENKd4Rp7xO4V
bowlexpkgRsZ9yLYpj+5AQ0EXu0m0wEIAL2J8jKMWlvbDdM+5KqRx2qGQWS4qXFy
xBWBfJCkBQdR7DrFA3qL/vPOrqfMGVNLQzVo84Edoa8DzbVARDhkC+CNu8Bmm2us
1rq6+f82CVxUoLHCJ7Y5AyFaX9itr8izz1poH2+/KJfqsWhVDQtvEWzCNZF23s+j
1F3C26phbXGGhSpz6ct487cG1IaJ8gExOIdkxZlRDfeniVcqt9qgFAX2lK9qZGz3
5v1ihttX+EGgKqs5Zh1ukAmJKqxcJXmJv2+4xmR6MLZNrLL2MGj5NyiX0BwxBcgH
ZQE/QEzJYdZi+Yf+Uv4N6yTavJAlAGsbwIJs1pvrIbBv67PVfZZfOmUAEQEAAYkB
NgQYAQgAIAIbDBYhBEkaBU4NHL/XNxtWU8Cg7XNtEc1bBQJe7ScCAAoJEMCg7XNt
Ec1bGqAIAIZA6temDDkVZP0Q6px6eLm8znPWndj1MIs6Cg6wUBHmU5L29t4If/HQ
SpNrlzQNTmOFAtxGxhPZZpWuCc1Y+N75tjLDWSZLpyVLIXRbf9n8YBkTlrwbqPbk
5NfkiPl/spd4UBM+RHIYQbxwWKdFJwYsWWEe5C4CikCn4l7gnfPSSC6JypbFP5Wf
Z6xvGUy8L2ModpwY3WcPsmKT/bCFURqWNtoMAaOu6uNjMV09LmyFBkCR5sxt+h5i
1XhwvC+eP5MKCJCbaYJlxmZxx9Ccj80XJxeaZCk+VakEu7fqApGwyKNSNWNhKDVs
NZB+32UzFDQaOCVCv6etmuY0b/Q0veg=
=yYbc
-----END PGP PUBLIC KEY BLOCK-----`
}

// privatekey returns a ASN.1 DER encoded key suitable for testing.
func privateKey() string {
	return `-----BEGIN PRIVATE KEY-----
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
ImXgKS5UmJB53s9yiomen/n7cUXvrAk=
-----END PRIVATE KEY-----`
}

// certificate returns a ASN.1 DER encoded certificate for the private key suitable for testing.
func certificate() string {
	return `-----BEGIN CERTIFICATE-----
MIIE6DCCAtACCQCzHO2a8qU6KzANBgkqhkiG9w0BAQsFADA2MRIwEAYDVQQDDAls
b2NhbGhvc3QxIDAeBgNVBAoMF0NsaWVudCBDZXJ0aWZpY2F0ZSBEZW1vMB4XDTE5
MTIwNTE3MjY1N1oXDTIwMTIwNDE3MjY1N1owNjESMBAGA1UEAwwJbG9jYWxob3N0
MSAwHgYDVQQKDBdDbGllbnQgQ2VydGlmaWNhdGUgRGVtbzCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBAJ3iM9y8yWa9P+MRVj9Jj+rpNGDDrS1z2kl7SgDB
ZisFYCHccvLfnz10OCWkI4fUWeAbsg2QJZ8i1EdFDtWnqhPXyqUsHRZRAI92ujvf
4g9r7ID0MENfVKOIY7ye3aphHUEW5lAkkJae12QcxyjTbrbfE0kQXpBHYaqFPRZs
CTXZOOZfSuEVxcrRJMjlt1J8thnGGmaa3yH6o1yt6hQdTp9JNzzeUwWe78PaAms8
RaLcaDXC7fsgByI8j9coKOluQdinYxkBLAHMpo7x6QuOYA3oguZXAUpMXJMJCQD2
Nd0WL33Dy4XHcPrhw+lfCVW7E1sbWO3Ka1ZVeu0FEF0erOKAmfQk2eoMJzR9qKwm
kbM2rYq5qrN19a5Rdtxxov4zOuOyvI6b1Uz5PajN1dyXzuKImOXFmiEL7ykC8kxD
u8HA90pC2VK3V0mx6tsR/H6zMBIg2je51nJ/11VCmIgS5/+jso1h+oUtqHAsWi1/
5u8lrQzMC3CR3VKLGCWhfF7NpQ82DYLnBh60t/E4mY0WX7GDAY8QTMKd4dnmKU6d
nKYDzXZR1he1c08+6NX+pdzJxih8Q/EG0PkNNla0MabMDsi7eFMUCjSPOUG99vGW
PNvMqP/EvCmCW7VKmph4NSNHjwkxTTOQD/ZGX9IpQWZkCxyCIMxAi1hCgu0zKR++
U+7BAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAC8av9I4ezwlmM7ysaJvC1IfCzNN
VawIK1U7bfj9Oyjl49Bn/yTwbbiQ8j5VjOza4umIwnYp1HP6/mlBO+ey8WFYPmDM
JAspk6sYEQW7MrbZ9QOmq24YAkwMzgK1hDASCKq4GJCzGDym3Zx6fvPnMCPdei2c
jgtjzzBmyewE0zcegOHDrFXTaUIfoSbduTbV9zClJ/pJDkTklRX1cYBtIox77gpZ
1cnIC803gi1rVCHRNdq85ltOTjoF1/wVamLy5c6CYlp5IPyVOm0nqbqra47QIwss
QSGxn5l52BC1jP1l3eK1mEr64+dbMhqX3ZQwhfuiQ9VmdovNN1NarPWfmQia6Spq
TfxN+3VhloKFUh+fgwNzWYLKCMnwBuPVhVGcpclvrj50MsyeiT2IfE8pqWw26g6g
0xu85AbqYKePaZ7wPoDddbwCIGr6BBT87Nsu+AqtnkH3uw34FDDcjWR1CmNuI1mP
ac6d1jdfbkL5ZUJTpTJi0BxWbTGupv8VzufteFRNa7U2h1O6+kyPmEpA3heEZcEO
Hq5zIfW6QTUmBXDfMFzQ9h0764oBVwm29bjZ59bU3RhcAZtL8fi5BapNtoKAy55d
P/0WahbwNjP68QYVLBeK9Sfo0XxLU0hJP4RJUZSXy9kUuZ8xhAM/6PdE04cDq71v
Zfq6/HA3phy85qyj
-----END CERTIFICATE-----`
}

// caCert returns a CA certificate suitable for testing
func caCert() string {
	return `-----BEGIN CERTIFICATE-----
MIICUTCCAfugAwIBAgIBADANBgkqhkiG9w0BAQQFADBXMQswCQYDVQQGEwJDTjEL
MAkGA1UECBMCUE4xCzAJBgNVBAcTAkNOMQswCQYDVQQKEwJPTjELMAkGA1UECxMC
VU4xFDASBgNVBAMTC0hlcm9uZyBZYW5nMB4XDTA1MDcxNTIxMTk0N1oXDTA1MDgx
NDIxMTk0N1owVzELMAkGA1UEBhMCQ04xCzAJBgNVBAgTAlBOMQswCQYDVQQHEwJD
TjELMAkGA1UEChMCT04xCzAJBgNVBAsTAlVOMRQwEgYDVQQDEwtIZXJvbmcgWWFu
ZzBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQCp5hnG7ogBhtlynpOS21cBewKE/B7j
V14qeyslnr26xZUsSVko36ZnhiaO/zbMOoRcKK9vEcgMtcLFuQTWDl3RAgMBAAGj
gbEwga4wHQYDVR0OBBYEFFXI70krXeQDxZgbaCQoR4jUDncEMH8GA1UdIwR4MHaA
FFXI70krXeQDxZgbaCQoR4jUDncEoVukWTBXMQswCQYDVQQGEwJDTjELMAkGA1UE
CBMCUE4xCzAJBgNVBAcTAkNOMQswCQYDVQQKEwJPTjELMAkGA1UECxMCVU4xFDAS
BgNVBAMTC0hlcm9uZyBZYW5nggEAMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEE
BQADQQA/ugzBrjjK9jcWnDVfGHlk3icNRq0oV7Ri32z/+HQX67aRfgZu7KWdI+Ju
Wm7DCfrPNGVwFWUQOmsPue9rZBgO
-----END CERTIFICATE-----`
}

func appendNewLine(s string) string {
	return s + "\n"
}

// escapePercentSign uses Terraform's escape syntax (i.e., repeating characters)
// to properly escape percent signs (i.e., '%').
//
// There are two significant places where '%' can show up:
// 1. Before a left curly brace (i.e., '{').
// 2. Not before a left curly brace.
//
// In case #1, we have to double escape so that Terraform does not interpret Fastly's
// configuration values as its own (e.g., https://docs.fastly.com/en/guides/custom-log-formats).
//
// In case #2, we only have to single escape.
//
// Refer to the Terraform documentation on string literals for more details:
// https://www.terraform.io/docs/configuration/expressions.html#string-literals
func escapePercentSign(s string) string {
	escapeSign := strings.ReplaceAll(s, "%", "%%")
	escapeTemplateSequence := strings.ReplaceAll(escapeSign, "%%{", "%%%%{")

	return escapeTemplateSequence
}

func TestEscapePercentSign(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "string no percent signs should change nothing",
			input: "abc 123",
			want:  "abc 123",
		},
		{
			name:  "one percent sign should return two percent signs",
			input: "%",
			want:  "%%",
		},
		{
			name:  "one percent sign mid-string should return two percent signs in the same place",
			input: "abc%123",
			want:  "abc%%123",
		},
		{
			name:  "one percent sign before left curly brace should return four percent signs then curly brace",
			input: "%{",
			want:  "%%%%{",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got := escapePercentSign(testcase.input)

			if got != testcase.want {
				t.Errorf("escapePercentSign(%q): \n\tgot: '%+v'\n\twant: '%+v'", testcase.input, got, testcase.want)
			}
		})
	}
}
