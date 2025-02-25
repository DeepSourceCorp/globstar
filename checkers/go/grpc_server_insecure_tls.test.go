func unsafe() {
	// <expect-error> insecure grpc tls server
	s := grpc.NewServer()
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}


func safe() {
	// Safe secure grpc tls server
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{})))
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}