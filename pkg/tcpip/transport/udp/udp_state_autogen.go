// automatically generated by stateify.

package udp

import (
	"gvisor.dev/gvisor/pkg/state"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
)

func (x *udpPacket) StateTypeName() string {
	return "pkg/tcpip/transport/udp.udpPacket"
}

func (x *udpPacket) StateFields() []string {
	return []string{
		"udpPacketEntry",
		"senderAddress",
		"packetInfo",
		"data",
		"timestamp",
		"tos",
	}
}

func (x *udpPacket) beforeSave() {}

func (x *udpPacket) StateSave(m state.Sink) {
	x.beforeSave()
	var data buffer.VectorisedView = x.saveData()
	m.SaveValue(3, data)
	m.Save(0, &x.udpPacketEntry)
	m.Save(1, &x.senderAddress)
	m.Save(2, &x.packetInfo)
	m.Save(4, &x.timestamp)
	m.Save(5, &x.tos)
}

func (x *udpPacket) afterLoad() {}

func (x *udpPacket) StateLoad(m state.Source) {
	m.Load(0, &x.udpPacketEntry)
	m.Load(1, &x.senderAddress)
	m.Load(2, &x.packetInfo)
	m.Load(4, &x.timestamp)
	m.Load(5, &x.tos)
	m.LoadValue(3, new(buffer.VectorisedView), func(y interface{}) { x.loadData(y.(buffer.VectorisedView)) })
}

func (x *endpoint) StateTypeName() string {
	return "pkg/tcpip/transport/udp.endpoint"
}

func (x *endpoint) StateFields() []string {
	return []string{
		"TransportEndpointInfo",
		"waiterQueue",
		"uniqueID",
		"rcvReady",
		"rcvList",
		"rcvBufSizeMax",
		"rcvBufSize",
		"rcvClosed",
		"sndBufSize",
		"sndBufSizeMax",
		"state",
		"dstPort",
		"v6only",
		"ttl",
		"multicastTTL",
		"multicastAddr",
		"multicastNICID",
		"multicastLoop",
		"portFlags",
		"bindToDevice",
		"broadcast",
		"noChecksum",
		"lastError",
		"boundBindToDevice",
		"boundPortFlags",
		"sendTOS",
		"receiveTOS",
		"receiveTClass",
		"receiveIPPacketInfo",
		"shutdownFlags",
		"multicastMemberships",
		"effectiveNetProtos",
		"owner",
		"linger",
	}
}

func (x *endpoint) StateSave(m state.Sink) {
	x.beforeSave()
	var rcvBufSizeMax int = x.saveRcvBufSizeMax()
	m.SaveValue(5, rcvBufSizeMax)
	var lastError string = x.saveLastError()
	m.SaveValue(22, lastError)
	m.Save(0, &x.TransportEndpointInfo)
	m.Save(1, &x.waiterQueue)
	m.Save(2, &x.uniqueID)
	m.Save(3, &x.rcvReady)
	m.Save(4, &x.rcvList)
	m.Save(6, &x.rcvBufSize)
	m.Save(7, &x.rcvClosed)
	m.Save(8, &x.sndBufSize)
	m.Save(9, &x.sndBufSizeMax)
	m.Save(10, &x.state)
	m.Save(11, &x.dstPort)
	m.Save(12, &x.v6only)
	m.Save(13, &x.ttl)
	m.Save(14, &x.multicastTTL)
	m.Save(15, &x.multicastAddr)
	m.Save(16, &x.multicastNICID)
	m.Save(17, &x.multicastLoop)
	m.Save(18, &x.portFlags)
	m.Save(19, &x.bindToDevice)
	m.Save(20, &x.broadcast)
	m.Save(21, &x.noChecksum)
	m.Save(23, &x.boundBindToDevice)
	m.Save(24, &x.boundPortFlags)
	m.Save(25, &x.sendTOS)
	m.Save(26, &x.receiveTOS)
	m.Save(27, &x.receiveTClass)
	m.Save(28, &x.receiveIPPacketInfo)
	m.Save(29, &x.shutdownFlags)
	m.Save(30, &x.multicastMemberships)
	m.Save(31, &x.effectiveNetProtos)
	m.Save(32, &x.owner)
	m.Save(33, &x.linger)
}

func (x *endpoint) StateLoad(m state.Source) {
	m.Load(0, &x.TransportEndpointInfo)
	m.Load(1, &x.waiterQueue)
	m.Load(2, &x.uniqueID)
	m.Load(3, &x.rcvReady)
	m.Load(4, &x.rcvList)
	m.Load(6, &x.rcvBufSize)
	m.Load(7, &x.rcvClosed)
	m.Load(8, &x.sndBufSize)
	m.Load(9, &x.sndBufSizeMax)
	m.Load(10, &x.state)
	m.Load(11, &x.dstPort)
	m.Load(12, &x.v6only)
	m.Load(13, &x.ttl)
	m.Load(14, &x.multicastTTL)
	m.Load(15, &x.multicastAddr)
	m.Load(16, &x.multicastNICID)
	m.Load(17, &x.multicastLoop)
	m.Load(18, &x.portFlags)
	m.Load(19, &x.bindToDevice)
	m.Load(20, &x.broadcast)
	m.Load(21, &x.noChecksum)
	m.Load(23, &x.boundBindToDevice)
	m.Load(24, &x.boundPortFlags)
	m.Load(25, &x.sendTOS)
	m.Load(26, &x.receiveTOS)
	m.Load(27, &x.receiveTClass)
	m.Load(28, &x.receiveIPPacketInfo)
	m.Load(29, &x.shutdownFlags)
	m.Load(30, &x.multicastMemberships)
	m.Load(31, &x.effectiveNetProtos)
	m.Load(32, &x.owner)
	m.Load(33, &x.linger)
	m.LoadValue(5, new(int), func(y interface{}) { x.loadRcvBufSizeMax(y.(int)) })
	m.LoadValue(22, new(string), func(y interface{}) { x.loadLastError(y.(string)) })
	m.AfterLoad(x.afterLoad)
}

func (x *multicastMembership) StateTypeName() string {
	return "pkg/tcpip/transport/udp.multicastMembership"
}

func (x *multicastMembership) StateFields() []string {
	return []string{
		"nicID",
		"multicastAddr",
	}
}

func (x *multicastMembership) beforeSave() {}

func (x *multicastMembership) StateSave(m state.Sink) {
	x.beforeSave()
	m.Save(0, &x.nicID)
	m.Save(1, &x.multicastAddr)
}

func (x *multicastMembership) afterLoad() {}

func (x *multicastMembership) StateLoad(m state.Source) {
	m.Load(0, &x.nicID)
	m.Load(1, &x.multicastAddr)
}

func (x *udpPacketList) StateTypeName() string {
	return "pkg/tcpip/transport/udp.udpPacketList"
}

func (x *udpPacketList) StateFields() []string {
	return []string{
		"head",
		"tail",
	}
}

func (x *udpPacketList) beforeSave() {}

func (x *udpPacketList) StateSave(m state.Sink) {
	x.beforeSave()
	m.Save(0, &x.head)
	m.Save(1, &x.tail)
}

func (x *udpPacketList) afterLoad() {}

func (x *udpPacketList) StateLoad(m state.Source) {
	m.Load(0, &x.head)
	m.Load(1, &x.tail)
}

func (x *udpPacketEntry) StateTypeName() string {
	return "pkg/tcpip/transport/udp.udpPacketEntry"
}

func (x *udpPacketEntry) StateFields() []string {
	return []string{
		"next",
		"prev",
	}
}

func (x *udpPacketEntry) beforeSave() {}

func (x *udpPacketEntry) StateSave(m state.Sink) {
	x.beforeSave()
	m.Save(0, &x.next)
	m.Save(1, &x.prev)
}

func (x *udpPacketEntry) afterLoad() {}

func (x *udpPacketEntry) StateLoad(m state.Source) {
	m.Load(0, &x.next)
	m.Load(1, &x.prev)
}

func init() {
	state.Register((*udpPacket)(nil))
	state.Register((*endpoint)(nil))
	state.Register((*multicastMembership)(nil))
	state.Register((*udpPacketList)(nil))
	state.Register((*udpPacketEntry)(nil))
}
