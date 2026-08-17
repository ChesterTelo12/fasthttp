--- client_test.go
+++ client_test.go
@@ -1815,3 +1815,65 @@
+
+func TestHostClientCloseIdleConnectionsRace(t *testing.T) {
+	ln, err := net.Listen("tcp", "127.0.0.1:0")
+	if err != nil {
+		t.Fatalf("cannot start listener: %s", err)
+	}
+	defer ln.Close()
+
+	ch := make(chan net.Conn, 10)
+	go func() {
+		for {
+			conn, err := ln.Accept()
+			if err != nil {
+				return
+			}
+			ch <- conn
+		}
+	}()
+
+	c := &HostClient{
+		Addr: ln.Addr().String(),
+	}
+
+	req := AcquireRequest()
+	req.SetRequestURI("http://" + c.Addr + "/")
+	resp := AcquireResponse()
+
+	go func() {
+		conn := <-ch
+		buf := make([]byte, 1024)
+		n, _ := conn.Read(buf)
+		if n > 0 {
+			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
+		}
+	}()
+
+	if err := c.Do(req, resp); err != nil {
+		t.Fatalf("unexpected error: %s", err)
+	}
+
+	done := make(chan struct{})
+	go func() {
+		c.CloseIdleConnections()
+		close(done)
+	}()
+
+	req2 := AcquireRequest()
+	req2.SetRequestURI("http://" + c.Addr + "/")
+	resp2 := AcquireResponse()
+
+	go func() {
+		conn := <-ch
+		buf := make([]byte, 1024)
+		n, _ := conn.Read(buf)
+		if n > 0 {
+			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
+		}
+	}()
+
+	if err := c.Do(req2, resp2); err != nil {
+		t.Fatalf("unexpected error: %s", err)
+	}
+
+	<-done
+}