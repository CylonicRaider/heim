package console

import (
	"bytes"
	"testing"

	"github.com/euphoria-io/scope"

	"euphoria.leet.nu/heim/backend/mock"
	"euphoria.leet.nu/heim/console"
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/security"
	"euphoria.leet.nu/heim/proto/snowflake"

	. "github.com/smartystreets/goconvey/convey"
)

type testPipes struct {
	input  bytes.Buffer
	output bytes.Buffer
}

func (t *testPipes) Read(out []byte) (int, error) {
	return t.input.Read(out)
}

func (t *testPipes) Write(in []byte) (int, error) {
	return t.output.Write(in)
}

func TestDeleteMessage(t *testing.T) {
	ctx := scope.New()
	kms := security.LocalKMS()
	kms.SetMasterKey(make([]byte, security.AES256.KeySize()))
	session := mock.TestSession("test", "T1", "ip1")

	runCommand := func(ctrl *Controller, argv []string) {
		pipes := &testPipes{}
		cli := newCLI(ctrl)
		con := console.NewDefaultConsole(pipes).WithContext(ctx.Fork())
		console.LaunchCLI(con, cli.CLI, argv)
	}

	sendMessage := func(room proto.Room) (proto.Message, error) {
		msgid, err := snowflake.New()
		if err != nil {
			return proto.Message{}, err
		}

		msg := proto.Message{
			ID: msgid,
			Sender: proto.SessionView{
				SessionID:    "test",
				IdentityView: proto.IdentityView{ID: "test"},
			},
			Content: "test",
		}

		if managedRoom, ok := room.(proto.ManagedRoom); ok {
			key, err := managedRoom.MessageKey(ctx)
			if err != nil {
				return proto.Message{}, err
			}

			if key != nil {
				mkey := key.ManagedKey()
				if err := kms.DecryptKey(&mkey); err != nil {
					return proto.Message{}, err
				}
				if err := proto.EncryptMessage(&msg, key.KeyID(), &mkey); err != nil {
					return proto.Message{}, err
				}
			}
		}

		result, err := room.Send(ctx, session, msg)
		return result, err
	}

	Convey("Delete message in public room", t, func() {
		ctrl := &Controller{
			backend: &mock.TestBackend{},
			kms:     kms,
			ctx:     ctx,
		}

		public, err := ctrl.backend.CreateRoom(ctx, kms, false, "public")
		So(err, ShouldBeNil)
		sent, err := sendMessage(public)
		So(err, ShouldBeNil)

		runCommand(ctrl, []string{"delete-message", "public:" + sent.ID.String()})

		deleted, err := public.GetMessage(ctx, sent.ID)
		So(deleted, ShouldBeNil)
		So(err, ShouldEqual, proto.ErrMessageNotFound)
	})

	Convey("Delete message in private room", t, func() {
		ctrl := &Controller{
			backend: &mock.TestBackend{},
			kms:     kms,
			ctx:     ctx,
		}

		private, err := ctrl.backend.CreateRoom(ctx, kms, true, "private")
		So(err, ShouldBeNil)
		runCommand(ctrl, []string{"lock-room", "private"})

		sent, err := sendMessage(private)
		So(err, ShouldBeNil)

		runCommand(ctrl, []string{"delete-message", "private:" + sent.ID.String()})

		deleted, err := private.GetMessage(ctx, sent.ID)
		So(deleted, ShouldBeNil)
		So(err, ShouldEqual, proto.ErrMessageNotFound)
	})
}
