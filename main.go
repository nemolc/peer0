package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server()
}

const CUSTOM_ID0 protocol.ID = "from_nemo0"

const CUSTOM_ID1 protocol.ID = "from_free"

func server() {
	ctx := context.Background()
	host, err := libp2p.New(ctx,
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}

	peerInfo := peerstore.AddrInfo{
		ID:    host.ID(),
		Addrs: host.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("host_id:", host.ID())
	fmt.Println("host_addrs:", host.Addrs())
	fmt.Println("addr:", addrs[0])

	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/63729/p2p/QmfKJL7waQRSKWUqnr7uTtzYRfT2zzTUFc7eMzvmJxwZMD")
	if err != nil {
		panic(err)
	}
	addr_info, err := peerstore.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	if err := host.Connect(ctx, *addr_info); err != nil {
		panic(err)
	}

	Invoke(ctx, host, addr_info.ID)

	exit_ch := make(chan os.Signal, 1)
	signal.Notify(exit_ch, syscall.SIGINT, syscall.SIGTERM)
	<-exit_ch
	fmt.Println("Received signal, shutting down...")
	if err := host.Close(); err != nil {
		panic(err)
	}
}

func Invoke(ctx context.Context, host host.Host, p peerstore.ID) {
	s, err := host.NewStream(ctx, p, CUSTOM_ID0, CUSTOM_ID1)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	fmt.Println("id:", s.ID())
	fmt.Println("protocol:", s.Protocol())
	_, err = s.Write([]byte("这是一个测试"))
	if err != nil {
		panic(err)
	}
	s.SetProtocol(CUSTOM_ID1)

	conn := s.Conn()


	s, err = conn.NewStream()
	if err != nil {
		panic(err)
	}

	fmt.Println("id:", s.ID())
	fmt.Println("protocol:", s.Protocol())

	_, err = s.Write([]byte("this is a test"))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
}

func serverPing(ch_in chan string) {
	// create a background context (i.e. one that never cancels)
	ctx := context.Background()

	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
	node, err := libp2p.New(ctx,
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}

	// configure our own ping protocol
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, Hello(pingService.PingHandler))

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("libp2p node address:", addrs[0])
	ch_in <- addrs[0].String()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")
	if err := node.Close(); err != nil {
		panic(err)
	}
}

func Hello(do func(s network.Stream)) func(s network.Stream) {
	return func(s network.Stream) {
		fmt.Println("id:", s.ID(), "protocol:", s.Protocol(), "stat:", s.Stat())
		do(s)
	}
}
