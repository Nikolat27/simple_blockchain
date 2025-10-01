package resolver

var dnsSeeds = []string{
	"127.0.0.1:8080",
}


func ResolveSeedNodes() []string {
	return dnsSeeds
}
