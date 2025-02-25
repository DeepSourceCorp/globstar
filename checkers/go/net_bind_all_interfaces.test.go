func bind_all() {
	// <expect-error> Bind to all interfaces
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
}

func bind_all2() {
	// <expect-error> Bind to all interfaces
	l, err := net.Listen("tcp", "0.0.0.0:2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
}

func bind_all3() {
	// Safe localhost
	l, err := net.Listen("tcp", "127.0.0.1:2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
}
