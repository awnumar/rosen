import socks

s = socks.socksocket()
s.set_proxy(socks.SOCKS5, "127.0.0.1", 23579)
s.connect(("127.0.0.1", 22222))

while True:
    data = input("> ")
    s.sendall(data.encode())
    print(s.recv(4096))
