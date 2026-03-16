package saiserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"



	saipb "github.com/openconfig/lemming/dataplane/proto/sai"
	fwdpb "github.com/openconfig/lemming/proto/forwarding"
)

func TestSetAclEntryAttributeFixed(t *testing.T) {
	dplane := &fakeSwitchDataplane{}
	c, a, stopFn := newTestACL(t, dplane)
	defer stopFn()
	a.mgr.StoreAttributes(a.mgr.NextID(), &saipb.SwitchAttribute{
		CpuPort: proto.Uint64(10),
	})

	a.tableToLocation[2] = tableLocation{
		groupID: "group1",
		bank:    0,
	}

	entryReq := &saipb.CreateAclEntryRequest{
		TableId:            proto.Uint64(2),
		ActionPacketAction: &saipb.AclActionData{Parameter: &saipb.AclActionData_PacketAction{PacketAction: saipb.PacketAction_PACKET_ACTION_DROP}},
	}
	a.mgr.StoreAttributes(100, entryReq)

	setReq := &saipb.SetAclEntryAttributeRequest{
		Oid: 100,
		ActionPacketAction: &saipb.AclActionData{Parameter: &saipb.AclActionData_PacketAction{PacketAction: saipb.PacketAction_PACKET_ACTION_FORWARD}},
	}

	_, err := c.SetAclEntryAttribute(context.Background(), setReq)
	if err != nil {
		t.Fatal(err)
	}
	
	if len(dplane.gotEntryRemoveReqs) == 0 {
		t.Fatalf("SetAclEntryAttribute() did not remove old entry")
	}
	wantRem := &fwdpb.TableEntryRemoveRequest{
		ContextId: &fwdpb.ContextId{Id: "foo"},
		TableId:   &fwdpb.TableId{ObjectId: &fwdpb.ObjectId{Id: "group1"}},
		EntryDesc: &fwdpb.EntryDesc{
			Entry: &fwdpb.EntryDesc_Flow{
				Flow: &fwdpb.FlowEntryDesc{
					Id:       100,
					Priority: 4294967295,
					Bank:     0,
				},
			},
		},
	}
	
	if diff := cmp.Diff(dplane.gotEntryRemoveReqs[0], wantRem, protocmp.Transform()); diff != "" {
		t.Errorf("SetAclEntryAttribute() unexpected remove request diff (-got +want):\n%s", diff)
	}
}
