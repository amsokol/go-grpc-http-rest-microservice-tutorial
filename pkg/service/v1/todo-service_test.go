package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/amsokol/go-grpc-http-rest-microservice-tutorial/pkg/api/v1"
)

func Test_toDoServiceServer_Create(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)
	tm := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(tm)

	type args struct {
		ctx context.Context
		req *v1.CreateRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.CreateResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO ToDo").WithArgs("title", "description", tm).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want: &v1.CreateResponse{
				Api: "v1",
				Id:  1,
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1000",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						Reminder: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
						},
					},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "Invalid Reminder field format",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						Reminder: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
						},
					},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "INSERT failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO ToDo").WithArgs("title", "description", tm).
					WillReturnError(errors.New("INSERT failed"))
			},
			wantErr: true,
		},
		{
			name: "LastInsertId failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.CreateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Title:       "title",
						Description: "description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO ToDo").WithArgs("title", "description", tm).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("LastInsertId failed")))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Create(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toDoServiceServer_Read(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)
	tm := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(tm)

	type args struct {
		ctx context.Context
		req *v1.ReadRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.ReadResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				rows := sqlmock.NewRows([]string{"ID", "Title", "Description", "Reminder"}).
					AddRow(1, "title", "description", tm)
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WithArgs(1).WillReturnRows(rows)
			},
			want: &v1.ReadResponse{
				Api: "v1",
				ToDo: &v1.ToDo{
					Id:          1,
					Title:       "title",
					Description: "description",
					Reminder:    reminder,
				},
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "SELECT failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WithArgs(1).
					WillReturnError(errors.New("SELECT failed"))
			},
			wantErr: true,
		},
		{
			name: "Not found",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				rows := sqlmock.NewRows([]string{"ID", "Title", "Description", "Reminder"})
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WithArgs(1).WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Read(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toDoServiceServer_Update(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)
	tm := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(tm)

	type args struct {
		ctx context.Context
		req *v1.UpdateRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.UpdateResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", tm, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want: &v1.UpdateResponse{
				Api:     "v1",
				Updated: 1,
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Reminder:    reminder,
					},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "Invalid Reminder field format",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Reminder: &timestamp.Timestamp{
							Seconds: 1,
							Nanos:   -1,
						},
					},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "UPDATE failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", tm, 1).
					WillReturnError(errors.New("UPDATE failed"))
			},
			wantErr: true,
		},
		{
			name: "RowsAffected failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", tm, 1).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("RowsAffected failed")))
			},
			wantErr: true,
		},
		{
			name: "Not Found",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.UpdateRequest{
					Api: "v1",
					ToDo: &v1.ToDo{
						Id:          1,
						Title:       "new title",
						Description: "new description",
						Reminder:    reminder,
					},
				},
			},
			mock: func() {
				mock.ExpectExec("UPDATE ToDo").WithArgs("new title", "new description", tm, 1).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Update(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toDoServiceServer_Delete(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)

	type args struct {
		ctx context.Context
		req *v1.DeleteRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.DeleteResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want: &v1.DeleteResponse{
				Api:     "v1",
				Deleted: 1,
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "DELETE failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).
					WillReturnError(errors.New("DELETE failed"))
			},
			wantErr: true,
		},
		{
			name: "RowsAffected failed",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("RowsAffected failed")))
			},
			wantErr: true,
		},
		{
			name: "Not Found",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.DeleteRequest{
					Api: "v1",
					Id:  1,
				},
			},
			mock: func() {
				mock.ExpectExec("DELETE FROM ToDo").WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.Delete(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toDoServiceServer_ReadAll(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := NewToDoServiceServer(db)
	tm1 := time.Now().In(time.UTC)
	reminder1, _ := ptypes.TimestampProto(tm1)
	tm2 := time.Now().In(time.UTC)
	reminder2, _ := ptypes.TimestampProto(tm2)

	type args struct {
		ctx context.Context
		req *v1.ReadAllRequest
	}
	tests := []struct {
		name    string
		s       v1.ToDoServiceServer
		args    args
		mock    func()
		want    *v1.ReadAllResponse
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadAllRequest{
					Api: "v1",
				},
			},
			mock: func() {
				rows := sqlmock.NewRows([]string{"ID", "Title", "Description", "Reminder"}).
					AddRow(1, "title 1", "description 1", tm1).
					AddRow(2, "title 2", "description 2", tm2)
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WillReturnRows(rows)
			},
			want: &v1.ReadAllResponse{
				Api: "v1",
				ToDos: []*v1.ToDo{
					{
						Id:          1,
						Title:       "title 1",
						Description: "description 1",
						Reminder:    reminder1,
					},
					{
						Id:          2,
						Title:       "title 2",
						Description: "description 2",
						Reminder:    reminder2,
					},
				},
			},
		},
		{
			name: "Empty",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadAllRequest{
					Api: "v1",
				},
			},
			mock: func() {
				rows := sqlmock.NewRows([]string{"ID", "Title", "Description", "Reminder"})
				mock.ExpectQuery("SELECT (.+) FROM ToDo").WillReturnRows(rows)
			},
			want: &v1.ReadAllResponse{
				Api:   "v1",
				ToDos: []*v1.ToDo{},
			},
		},
		{
			name: "Unsupported API",
			s:    s,
			args: args{
				ctx: ctx,
				req: &v1.ReadAllRequest{
					Api: "v1",
				},
			},
			mock:    func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.ReadAll(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("toDoServiceServer.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDoServiceServer.ReadAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
