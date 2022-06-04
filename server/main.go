package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pb "github.com/tou-tou/realtime-grpc/proto/world"
)

var (
	//tls  = flag.Bool("tls", false, "Connection users TLS if true, else plain TCP")
	port = flag.Int("port", 50051, "The server port")
)

// implement room interface
type roomServer struct {
	// key is room_id , value is array of user_id
	rooms map[string][]string
	// key is user_id , value is user info
	users map[string]*pb.User

	pb.UnimplementedRoomServer
}

func (s *roomServer) isExistUser(roomID string, userID string) bool {
	userIDs, isExist := s.rooms[roomID]
	if !isExist {
		return false
	}
	for _, v := range userIDs {
		if v == userID {
			return true
		}
	}
	return false
}

func (s *roomServer) roomUsers(roomID string) []*pb.User {
	var users []*pb.User
	for _, userID := range s.rooms[roomID] {
		users = append(users, s.users[userID])
	}
	return users
}

func indexOfArray(s []string, userID string) int {
	for i, v := range s {
		if v == userID {
			return i
		}
	}
	return -1
}

// UpdateTransform is update user's transform on user list
func (s *roomServer) UpdateTransform(userID string, user *pb.User) {
	origin := user.GetOrigin()
	s.users[userID].Origin = &pb.Transform{
		Pos: &pb.Position{
			X: origin.GetPos().GetX(),
			Y: origin.GetPos().GetY(),
			Z: origin.GetPos().GetZ(),
		},
		Rot: &pb.EulerRotation{
			X: origin.GetRot().GetX(),
			Y: origin.GetRot().GetY(),
			Z: origin.GetRot().GetZ(),
		},
	}

	head := user.GetHead()
	s.users[userID].Head = &pb.Transform{
		Pos: &pb.Position{
			X: head.GetPos().GetX(),
			Y: head.GetPos().GetY(),
			Z: head.GetPos().GetZ(),
		},
		Rot: &pb.EulerRotation{
			X: head.GetRot().GetX(),
			Y: head.GetRot().GetY(),
			Z: head.GetRot().GetZ(),
		},
	}

	left := user.GetLeftHand()
	s.users[userID].LeftHand = &pb.Transform{
		Pos: &pb.Position{
			X: left.GetPos().GetX(),
			Y: left.GetPos().GetY(),
			Z: left.GetPos().GetZ(),
		},
		Rot: &pb.EulerRotation{
			X: left.GetRot().GetX(),
			Y: left.GetRot().GetY(),
			Z: left.GetRot().GetZ(),
		},
	}

	right := user.GetRightHand()
	s.users[userID].RightHand = &pb.Transform{
		Pos: &pb.Position{
			X: right.GetPos().GetX(),
			Y: right.GetPos().GetY(),
			Z: right.GetPos().GetZ(),
		},
		Rot: &pb.EulerRotation{
			X: right.GetRot().GetX(),
			Y: right.GetRot().GetY(),
			Z: right.GetRot().GetZ(),
		},
	}
}

// Join function requires room_id and return user_id
// register user to room
func (s *roomServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	roomID := req.GetRoomId()

	// check the room with requested room_id
	_, isExist := s.rooms[roomID]
	if !isExist {
		s := fmt.Sprintf("room with %s does not exit", roomID)
		return nil, errors.New(s)
	}

	// generate user_id
	userUUID, _ := uuid.NewRandom()
	userID := userUUID.String()
	m := fmt.Sprintf("join! world_id:%s user_id:%s", roomID, userID)
	fmt.Println(m)

	// add user_id to room's user list
	s.rooms[roomID] = append(s.rooms[roomID], userID)
	// add user info to users
	s.users[userID] = &pb.User{
		UserId:    userID,
		Origin:    &pb.Transform{},
		Head:      &pb.Transform{},
		LeftHand:  &pb.Transform{},
		RightHand: &pb.Transform{},
	}
	// return user_id
	return &pb.JoinResponse{
		UserId: userID,
	}, nil

}

func (s *roomServer) Sync(stream pb.Room_SyncServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		roomID := in.GetRoomId()
		userID := in.GetUser().UserId

		// room is not exist
		_, isExistRoom := s.rooms[roomID]
		if !isExistRoom {
			s := fmt.Sprintf("room with %s does not exit", roomID)
			return errors.New(s)
		}
		// user is not in room
		if !s.isExistUser(roomID, userID) {
			s := fmt.Sprintf("user with %s does not exit", userID)
			return errors.New(s)
		}

		// update user's pos and rot
		user := in.GetUser()
		s.UpdateTransform(userID, user)

		if err := stream.Send(&pb.SyncResponse{
			Users: s.roomUsers(roomID),
		}); err != nil {
			return err
		}
	}
}

func (s *roomServer) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	roomID := req.GetRoomId()
	userID := req.GetUserId()
	// room is not exist
	_, isExistRoom := s.rooms[roomID]
	if !isExistRoom {
		s := fmt.Sprintf("room with %s does not exit", roomID)
		return nil, errors.New(s)
	}
	// user is not in room
	if !s.isExistUser(roomID, userID) {
		s := fmt.Sprintf("user with %s does not exit", userID)
		return nil, errors.New(s)
	}

	m := fmt.Sprintf("Leave! world_id:%s user_id:%s", roomID, userID)
	fmt.Println(m)

	// delete user info from room and user list
	index := indexOfArray(s.rooms[roomID], userID)
	userArr := s.rooms[roomID]
	s.rooms[roomID][index] = userArr[len(userArr)-1]
	s.rooms[roomID] = userArr[:len(userArr)-1]
	delete(s.users, userID)

	return &pb.LeaveResponse{}, nil
}

func main() {
	// Parse parses the command-line flags from os.Args[1:].
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen :%v", err)
	}
	grpcServer := grpc.NewServer()
	fmt.Println("Start gRPC Server")
	pb.RegisterRoomServer(grpcServer, &roomServer{
		rooms:                   map[string][]string{"metaverse": {}}, // initial room_id is only "world"
		users:                   map[string]*pb.User{},
		UnimplementedRoomServer: pb.UnimplementedRoomServer{},
	})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
