package credhub_test

import (
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/credhub-cli/credhub"
)

var _ = Describe("Client()", func() {
	It("should return a simple http.Client", func() {
		ch, _ := New("http://example.com", ServerVersion("2.2.2"))
		client := ch.Client()

		Expect(client).ToNot(BeNil())
	})

	Context("With ClientCert", func() {
		It("should return a http.Client with tls.Config with client cert", func() {
			ch, err := New("https://example.com", ClientCert("./fixtures/auth-tls-cert.pem", "./fixtures/auth-tls-key.pem"), ServerVersion("2.2.2"))
			Expect(err).NotTo(HaveOccurred())

			client := ch.Client()

			transport := client.Transport.(*http.Transport)
			tlsConfig := transport.TLSClientConfig
			Expect(len(tlsConfig.Certificates)).To(Equal(1))
			clientCert := tlsConfig.Certificates[0]
			x509Cert, err := x509.ParseCertificate(clientCert.Certificate[0])
			Expect(err).NotTo(HaveOccurred())

			Expect(x509Cert.Subject.CommonName).To(Equal("example.com"))
		})

		It("doesnt set any client certs if not used", func() {
			ch, err := New("https://example.com", ServerVersion("2.2.2"))
			Expect(err).NotTo(HaveOccurred())

			client := ch.Client()
			transport := client.Transport.(*http.Transport)
			tlsConfig := transport.TLSClientConfig
			Expect(tlsConfig.Certificates).To(BeEmpty())
		})

		It("fails creation with invalid cert,key pair", func() {
			_, err := New("https://example.com", ClientCert("./fixtures/auth-tls-key.pem", "./fixtures/auth-tls-cert.pem"), ServerVersion("2.2.2"))
			Expect(err).To(HaveOccurred())
		})
	})

	Context("With CaCerts", func() {
		It("should return a http.Client with tls.Config with RootCAs", func() {
			fixturePath := "./fixtures/"
			caCertFiles := []string{
				"auth-tls-ca.pem",
				"server-tls-ca.pem",
				"extra-ca.pem",
			}
			var caCerts []string
			expectedRootCAs := x509.NewCertPool()
			for _, caCertFile := range caCertFiles {
				caCertBytes, err := ioutil.ReadFile(fixturePath + caCertFile)
				if err != nil {
					Fail("Couldn't read certificate " + caCertFile + ": " + err.Error())
				}

				caCerts = append(caCerts, string(caCertBytes))
				expectedRootCAs.AppendCertsFromPEM(caCertBytes)
			}

			ch, _ := New("https://example.com", CaCerts(caCerts...), ServerVersion("2.2.2"))

			client := ch.Client()

			transport := client.Transport.(*http.Transport)
			tlsConfig := transport.TLSClientConfig

			Expect(client.Timeout).To(Equal(45 * time.Second))

			Expect(tlsConfig.InsecureSkipVerify).To(BeFalse())
			Expect(tlsConfig.PreferServerCipherSuites).To(BeTrue())
			Expect(tlsConfig.RootCAs.Subjects()).To(ConsistOf(expectedRootCAs.Subjects()))
		})
	})

	Context("With InsecureSkipVerify", func() {
		It("should return a http.Client with tls.Config without RootCAs", func() {
			ch, _ := New("https://example.com", SkipTLSValidation(true), ServerVersion("2.2.2"))
			client := ch.Client()

			transport := client.Transport.(*http.Transport)
			tlsConfig := transport.TLSClientConfig

			Expect(client.Timeout).To(Equal(45 * time.Second))

			Expect(tlsConfig.InsecureSkipVerify).To(BeTrue())
			Expect(tlsConfig.PreferServerCipherSuites).To(BeTrue())
		})
	})
})
