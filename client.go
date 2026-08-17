--- client.go
+++ client.go
@@ -1008,9 +1008,14 @@
 func (c *HostClient) CloseIdleConnections() {
 	c.connsLock.Lock()
-	conns := c.conns
-	c.conns = nil
+	var conns []net.Conn
+	if len(c.conns) > 0 {
+		conns = make([]net.Conn, len(c.conns))
+		for i, cc := range c.conns {
+			conns[i] = cc.c
+		}
+		c.conns = nil
+	}
 	c.connsLock.Unlock()
 
-	for _, cc := range conns {
-		cc.c.Close()
+	for _, conn := range conns {
+		conn.Close()
 	}
 }