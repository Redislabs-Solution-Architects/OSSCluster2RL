package main

import (
	"testing"

	"github.com/go-redis/redis"

	osscluster2rl "github.com/Redislabs-Solution-Architects/OSSCluster2RL/helpers"
)

var azureClusterInfo = `d1eacb222624c77b00117a9b5f44c14b92ebf274 23.101.193.229:13003 slave 6d6bdca99ac3c2de5b25ad0cd468965eaf0c0cef 0 1596469840000 1 connected
5d1ebaa61f5fd3849a05834e2f6e9ffa2bc38d19 23.101.193.229:13004 master - 0 1596469841000 3 connected 5462-8191 13653-16383
22e92ba178225f9e65201570869eb96e74844667 23.101.193.229:13005 slave 5d1ebaa61f5fd3849a05834e2f6e9ffa2bc38d19 0 1596469840428 4 connected
e6f6603c88b51a9de1878635a81e25e3d3a81eed 23.101.193.229:13001 slave aabbc2b9e51d15394c418c09144bddfde2bc1370 0 1596469840209 2 connected
6d6bdca99ac3c2de5b25ad0cd468965eaf0c0cef 23.101.193.229:13002 myself,master - 0 1596469841000 1 connected 8192-13652
aabbc2b9e51d15394c418c09144bddfde2bc1370 23.101.193.229:13000 master - 0 1596469841626 2 connected 0-5461`

var disconnectedSlaveInfo = `3f63ff407ec6af6b39b64f8363f2ca1b5fa8e774 192.168.0.1:30001@40001 myself,master - 0 0 0 connected
2491efec2cc955d5317f2d71277d9e80b10a3a39 :0@0 slave,fail,noaddr - 1574414519922 1574414519922 474 disconnected
`

var brokenClusterInfo = `3f63ff407ec6af6b39b64f8363f2ca1b5fa8e774 :30001@40001 myself,master - 0 0 0 connected
2a9b5e4aa049ac0186d5ad3b95109909ed9eba22 127.0.0.1:30006@40006 slave 3f63ff407ec6af6b39b64f8363f2ca1b5fa8e774 0 1593352361101 6 connected
`

var brokenClusterInfoSlave = `3f63ff407ec6af6b39b64f8363f2ca1b5fa8e774 127.0.0.1:30001@40001 myself,master - 0 0 0 connected
2a9b5e4aa049ac0186d5ad3b95109909ed9eba22 127.0.0.1:AAA@40006 slave 3f63ff407ec6af6b39b64f8363f2ca1b5fa8e774 0 1593352361101 6 connected
`

var simpleClusterInfo = `de95b8cbbfc208ddf4be64de7b811c652b6d380c 127.0.0.1:30004@40004 slave 75e57a15b9e8ebf97fc5eb5390b493e08537c823 0 1593352361101 4 connected
f59f3a29fc64553b618554abe000833c7d1ddf05 127.0.0.1:30005@40005 myself,slave 3741c6f80aff0795f0a51d991be6d68edd0478ea 0 1593352361000 5 connected
2a9b5e4aa049ac0186d5ad3b95109909ed9eba22 127.0.0.1:30006@40006 slave c8cf86e9dc35daac9588c21dad4b4676917de1ed 0 1593352361101 6 connected
75e57a15b9e8ebf97fc5eb5390b493e08537c823 127.0.0.1:30002@40002 master - 0 1593352361102 2 connected 5461-10922
3741c6f80aff0795f0a51d991be6d68edd0478ea 127.0.0.1:30003@40003 master - 0 1593352361101 3 connected 10923-16383
c8cf86e9dc35daac9588c21dad4b4676917de1ed 127.0.0.1:30001@40001 master - 0 1593352361101 1 connected 0-5460
`

var complexClusterInfo = `247a44e0b92dcdea10b2bff4320a664b41d89863 11.1.8.135:6379@16379 master - 0 1593218491000 949 connected 12567-12571 12573-13274 13276-13312
1d83ef752f3fb425419c3b1c3b43ae09b5b9c619 11.1.8.240:6379@16379 master - 0 1593218490575 951 connected 6424-7167
15c96a647d6c52488d8ed21cf6faf89bd05c0b49 11.1.8.171:6379@16379 master - 0 1593218490000 948 connected 274-1023 1294-1295
cdcb85f7235c895506f45b51948999454f9f0710 11.1.8.213:6379@16379 slave 5f069f8a114b8443dfe58ab6c09088d1fad27862 0 1593218490000 953 connected
feb161fb4cdea40a96d21ffdaddc75198b25fe0f 11.1.8.232:6379@16379 slave 65ba518c32c49f09e2055ff3e0a667bcc46d0eb7 0 1593218491587 945 connected
d04c33a9154839046c4628641c01f32df6fe567d 11.1.8.229:6379@16379 slave 4326d0470c5a76da648632a6f6fa3e95d6ed81fb 0 1593218491587 891 connected
ec3657d8a5c83d4ddd242c1ca3d49578c6262677 11.1.0.248:6379@16379 master - 0 1593218491588 884 connected 15640-16383
e6132e25809db05d968360ca16a98bc44f5c52cb 11.1.8.180:6379@16379 slave 2d5cbc75f634697fb6dccf61e3ebd4dfb4e38d56 0 1593218490575 926 connected
57d4aaf6e1d3b19f3945a253574b7d728c3d547e 11.1.0.243:6379@16379 slave 0e2778eb817ce14ec2257f0a30e5a61dbbec3bb0 0 1593218490000 933 connected
7fad753801ed9d9a0da2bb1903600acd4497f244 11.1.8.51:6379@16379 master - 0 1593218490000 919 connected 4376-5119
930b071c12a8a1e813545b4d99c45e6774bb65cb 11.1.8.231:6379@16379 master - 0 1593218491000 946 connected 1024-1283 14337-14339 14341-14617 15382-15585
c9b971126b2d2863a0c75544a6638c8def67000b 11.1.8.252:6379@16379 slave 0fb393d721410f2f80122b2f091c22c0768f4531 0 1593218491000 917 connected
5c2265c60cc25ddfe309c3c89bc0600a61a98cd4 11.1.0.162:6379@16379 slave 98c8ba6fd08e42ac843666a985f4cd88af509b5d 0 1593218491285 952 connected
8db2c596451eb44ad4d1a937e9c47b9b78c20bb2 11.1.8.244:6379@16379 master - 0 1593218491587 890 connected 7448-8191
604efc24a75eb522c7ff9569279ecc2b07e51322 11.1.0.210:6379@16379 master - 0 1593218491000 924 connected 10520-11263
64c3aa20ee49647bfe1b27e94c205ef3961dd9d3 11.1.0.200:6379@16379 slave 8e179b93a87c5017142d571d29826c3ff84cec31 0 1593218490576 950 connected
c3501f9fdebf1e394e5c7c867c26b1c13a6bcba1 11.1.0.173:6379@16379 slave cc3ccf5ed920422607b329c8b2a6ffd191452670 0 1593218490000 897 connected
20a5ed0f3f6a635bbd479c6a0990b8a9f4e9a91f 11.1.8.47:6379@16379 slave 64c8c4602cf9f15b1bd0430e36a94c21aad28786 0 1593218492093 887 connected
525a8f584aac32e6a34b645e4a1021a82ef459b5 11.1.0.177:6379@16379 slave 1d83ef752f3fb425419c3b1c3b43ae09b5b9c619 0 1593218490000 951 connected
45796d27078fe633fded7f26ae2ec9bb5c386dc2 11.1.0.45:6379@16379 master - 0 1593218491790 895 connected 11548-12287 12564 12572 13275 14340
98c8ba6fd08e42ac843666a985f4cd88af509b5d 11.1.8.92:6379@16379 master - 0 1593218491587 952 connected 9496-10239
64c8c4602cf9f15b1bd0430e36a94c21aad28786 11.1.0.144:6379@16379 master - 0 1593218492094 887 connected 2328-3071
d8a4b2a2e25dfd16ab670abafe877e035f49e1e5 11.1.8.136:6379@16379 master - 0 1593218490000 947 connected 0-119 1284-1293 5120-5399 7168-7447 15586-15639
2d5cbc75f634697fb6dccf61e3ebd4dfb4e38d56 11.1.8.79:6379@16379 master - 0 1593218491587 926 connected 2230-2327 3072-3229 4096-4303 13313-13592
d9202602ba0826b9a232a8cc94a1676cfd566cba 11.1.8.36:6379@16379 slave 4bd19995b29c4652026e22660cc7feae7f8aabc7 0 1593218491000 935 connected
15a8fdaef0b8545cd588da213a1b8edc49e06c8d 11.1.0.22:6379@16379 slave 15c96a647d6c52488d8ed21cf6faf89bd05c0b49 0 1593218491588 948 connected
62b13835d97d198248e282d8b5273e0b79599799 11.1.8.198:6379@16379 slave 5fbafba891519d12268402d685cdb93833bd4d2d 0 1593218490000 882 connected
790a242038b6897badc35f0a4a18b05ba3a5df67 11.1.8.192:6379@16379 slave d8a4b2a2e25dfd16ab670abafe877e035f49e1e5 0 1593218491000 947 connected
4326d0470c5a76da648632a6f6fa3e95d6ed81fb 11.1.0.82:6379@16379 master - 0 1593218490576 891 connected 14618-15361
5f069f8a114b8443dfe58ab6c09088d1fad27862 11.1.0.120:6379@16379 master - 0 1593218491588 953 connected 1296-2047
a5ba27fd4ba9944600f7a8b53f2ccee876cd3673 11.1.8.52:6379@16379 myself,slave 45796d27078fe633fded7f26ae2ec9bb5c386dc2 0 1593218490000 800 connected
65ba518c32c49f09e2055ff3e0a667bcc46d0eb7 11.1.8.185:6379@16379 master - 0 1593218490575 945 connected 120-273 6144-6423 8192-8471 9466-9495
0e2778eb817ce14ec2257f0a30e5a61dbbec3bb0 11.1.8.80:6379@16379 master - 0 1593218491284 933 connected 8472-9215
2820bc7dec9a617c1e5ec55ce9b6aeb7fb40f341 11.1.0.172:6379@16379 slave 8db2c596451eb44ad4d1a937e9c47b9b78c20bb2 0 1593218490576 890 connected
384113ea6d58f88de31cadc8e7c84880353022c7 11.1.0.150:6379@16379 slave 247a44e0b92dcdea10b2bff4320a664b41d89863 0 1593218491000 949 connected
05128cd4a87be01ae25005ccad8f01294003a431 11.1.0.133:6379@16379 slave 7fad753801ed9d9a0da2bb1903600acd4497f244 0 1593218491588 919 connected
5fbafba891519d12268402d685cdb93833bd4d2d 11.1.0.171:6379@16379 master - 0 1593218491588 882 connected 5400-6143
1963dd88db65fdbeb9f5c29973280c4a2176026d 11.1.8.45:6379@16379 slave ec3657d8a5c83d4ddd242c1ca3d49578c6262677 0 1593218490000 884 connected
0fb393d721410f2f80122b2f091c22c0768f4531 11.1.8.160:6379@16379 master - 0 1593218490000 917 connected 2048-2229 11264-11547 12288-12563 12565-12566
4bd19995b29c4652026e22660cc7feae7f8aabc7 11.1.8.190:6379@16379 master - 0 1593218491082 935 connected 3230-3351 4304-4375 9216-9465 10240-10519 15362-15381
b15d799f3f7959a10ba95077e5b3253162c1710d 11.1.8.9:6379@16379 slave 604efc24a75eb522c7ff9569279ecc2b07e51322 0 1593218490000 924 connected
cc3ccf5ed920422607b329c8b2a6ffd191452670 11.1.8.153:6379@16379 master - 0 1593218490000 897 connected 13593-14336
8e179b93a87c5017142d571d29826c3ff84cec31 11.1.8.165:6379@16379 master - 0 1593218489766 950 connected 3352-4095
fcc2a37a259b86bd2d9fb25880d8cde4404e7552 11.1.8.117:6379@16379 slave 930b071c12a8a1e813545b4d99c45e6774bb65cb 0 1593218490575 946 connected
`

// TestAzureParsing: Test Node parsing for Azure
func TestAzureParsing(t *testing.T) {
	j := redis.NewStringResult(azureClusterInfo, nil)
	f, err := osscluster2rl.ParseNodes(j)
	if err != nil {
		t.Error("This should not catch an error: ", err)
	}

	if len(f) < 2 {
		t.Error("This should return more than 2 nodes returned :", f)
	}
}

// TestParsingBroken : Test Node parsing
func TestParsingBroken(t *testing.T) {
	j := redis.NewStringResult(brokenClusterInfo, nil)
	_, err := osscluster2rl.ParseNodes(j)
	if err == nil {
		t.Error("This should catch an error")
	}
}

// TestParsingBrokenSlave : Test Node parsing
func TestParsingBrokenSlave(t *testing.T) {
	j := redis.NewStringResult(brokenClusterInfoSlave, nil)
	_, err := osscluster2rl.ParseNodes(j)
	if err == nil {
		t.Error("This should catch an error")
	}
}

// TestParsingDisconnectedSlave: Test Node parsing with a disconnected slave
func TestParsingDisconnectedSlave(t *testing.T) {
	j := redis.NewStringResult(disconnectedSlaveInfo, nil)
	q, err := osscluster2rl.ParseNodes(j)

	if len(q[0].Slaves) != 0 {
		t.Error("This should return an empty slave count got:", q[0])
	}

	if err != nil {
		t.Error("This should not catch an error:", err)
	}
}

// TestParsingSimple : Test Node parsing
func TestParsingSimple(t *testing.T) {
	j := redis.NewStringResult(simpleClusterInfo, nil)
	k, err := osscluster2rl.ParseNodes(j)
	if err != nil {
		t.Error("Errored: ", err)
	}
	if len(k) != 6 {
	}
}

// TestParsingComplex : Test Node parsing
func TestParsingComplex(t *testing.T) {
	j := redis.NewStringResult(complexClusterInfo, nil)
	k, err := osscluster2rl.ParseNodes(j)
	if err != nil {
		t.Error("Errored: ", err)
	}
	if len(k) != 44 {
		t.Error("Expected to find 6 servers, but got", len(k))
	}
}

// TestMastersSimple : Test Findind Node Masters
func TestMastersSimple(t *testing.T) {
	j := redis.NewStringResult(simpleClusterInfo, nil)
	k, _ := osscluster2rl.ParseNodes(j)
	m := osscluster2rl.ListMasters(k)
	if len(m) != 3 {
		t.Error("Expected to find 3 masters, but got", len(m))
	}
}

// TestMastersComplex : Test Findind Node Masters
func TestMastersComplex(t *testing.T) {
	j := redis.NewStringResult(complexClusterInfo, nil)
	k, _ := osscluster2rl.ParseNodes(j)
	m := osscluster2rl.ListMasters(k)
	if len(m) != 22 {
		t.Error("Expected to find 22 masters, but got", len(m))
	}
}
