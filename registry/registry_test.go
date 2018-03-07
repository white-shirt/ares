package registry

/*
func Test_Registry(t *testing.T) {
	r := NewETCDRegistry(
		WithAddrs([]string{"10.1.41.52:2379"}),
		WithTimeout(time.Second*10),
		WithSecure(false),
	)

	r.RegisterWithMetadata("test:v1:example", "10.1.41.52", metadata.New(map[string]string{
		"a": "1",
		"b": "2",
	}))

	r.RegisterWithMetadata("test:v1:example", "10.1.41.53", metadata.New(map[string]string{
		"b": "3",
	}))
}

func Test_Discovery(t *testing.T) {
	//cc, err := clientv3.New(
	//clientv3.Config{
	//Endpoints: []string{"10.1.41.52:2379"},
	//AutoSyncInterval: 0,
	//},
	//)

}
*/
