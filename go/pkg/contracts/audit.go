package contracts

import (
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
	commonv1 "github.com/ting-boundless/boundless/gen/go/ting/common/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuditToProto maps pkg/audit.Event to the generated proto message.
func AuditToProto(e audit.Event) (*commonv1.AuditEvent, error) {
	var data *structpb.Struct
	if len(e.Data) > 0 {
		var err error
		data, err = structpb.NewStruct(e.Data)
		if err != nil {
			return nil, err
		}
	}
	ts := e.Time
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return &commonv1.AuditEvent{
		Id:          e.ID,
		Source:      e.Source,
		Type:        e.Type,
		Subject:     e.Subject,
		Time:        timestamppb.New(ts),
		TenantId:    e.TenantID,
		ActorUserId: e.ActorUserID,
		Data:        data,
	}, nil
}

// AuditFromProto maps the generated proto message to pkg/audit.Event.
func AuditFromProto(m *commonv1.AuditEvent) audit.Event {
	if m == nil {
		return audit.Event{}
	}
	var data map[string]any
	if m.GetData() != nil {
		data = m.GetData().AsMap()
	}
	var ts time.Time
	if t := m.GetTime(); t != nil {
		ts = t.AsTime()
	}
	return audit.Event{
		ID:          m.GetId(),
		Source:      m.GetSource(),
		Type:        m.GetType(),
		Subject:     m.GetSubject(),
		Time:        ts,
		TenantID:    m.GetTenantId(),
		ActorUserID: m.GetActorUserId(),
		Data:        data,
	}
}
