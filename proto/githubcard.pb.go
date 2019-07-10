// Code generated by protoc-gen-go. DO NOT EDIT.
// source: githubcard.proto

package githubcard

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Issue_IssueState int32

const (
	Issue_OPEN   Issue_IssueState = 0
	Issue_CLOSED Issue_IssueState = 1
)

var Issue_IssueState_name = map[int32]string{
	0: "OPEN",
	1: "CLOSED",
}

var Issue_IssueState_value = map[string]int32{
	"OPEN":   0,
	"CLOSED": 1,
}

func (x Issue_IssueState) String() string {
	return proto.EnumName(Issue_IssueState_name, int32(x))
}

func (Issue_IssueState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{3, 0}
}

type Issue_Origin int32

const (
	Issue_UNKNOWN       Issue_Origin = 0
	Issue_FROM_WEB      Issue_Origin = 1
	Issue_FROM_RECEIVER Issue_Origin = 2
)

var Issue_Origin_name = map[int32]string{
	0: "UNKNOWN",
	1: "FROM_WEB",
	2: "FROM_RECEIVER",
}

var Issue_Origin_value = map[string]int32{
	"UNKNOWN":       0,
	"FROM_WEB":      1,
	"FROM_RECEIVER": 2,
}

func (x Issue_Origin) String() string {
	return proto.EnumName(Issue_Origin_name, int32(x))
}

func (Issue_Origin) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{3, 1}
}

type SilenceRequest_SilenceState int32

const (
	SilenceRequest_UNKNOWN   SilenceRequest_SilenceState = 0
	SilenceRequest_SILENCE   SilenceRequest_SilenceState = 1
	SilenceRequest_UNSILENCE SilenceRequest_SilenceState = 2
)

var SilenceRequest_SilenceState_name = map[int32]string{
	0: "UNKNOWN",
	1: "SILENCE",
	2: "UNSILENCE",
}

var SilenceRequest_SilenceState_value = map[string]int32{
	"UNKNOWN":   0,
	"SILENCE":   1,
	"UNSILENCE": 2,
}

func (x SilenceRequest_SilenceState) String() string {
	return proto.EnumName(SilenceRequest_SilenceState_name, int32(x))
}

func (SilenceRequest_SilenceState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{5, 0}
}

type Token struct {
	Token                string   `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Token) Reset()         { *m = Token{} }
func (m *Token) String() string { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()    {}
func (*Token) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{0}
}

func (m *Token) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Token.Unmarshal(m, b)
}
func (m *Token) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Token.Marshal(b, m, deterministic)
}
func (m *Token) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Token.Merge(m, src)
}
func (m *Token) XXX_Size() int {
	return xxx_messageInfo_Token.Size(m)
}
func (m *Token) XXX_DiscardUnknown() {
	xxx_messageInfo_Token.DiscardUnknown(m)
}

var xxx_messageInfo_Token proto.InternalMessageInfo

func (m *Token) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

type Silence struct {
	Silence              string   `protobuf:"bytes,1,opt,name=silence,proto3" json:"silence,omitempty"`
	Origin               string   `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Silence) Reset()         { *m = Silence{} }
func (m *Silence) String() string { return proto.CompactTextString(m) }
func (*Silence) ProtoMessage()    {}
func (*Silence) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{1}
}

func (m *Silence) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Silence.Unmarshal(m, b)
}
func (m *Silence) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Silence.Marshal(b, m, deterministic)
}
func (m *Silence) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Silence.Merge(m, src)
}
func (m *Silence) XXX_Size() int {
	return xxx_messageInfo_Silence.Size(m)
}
func (m *Silence) XXX_DiscardUnknown() {
	xxx_messageInfo_Silence.DiscardUnknown(m)
}

var xxx_messageInfo_Silence proto.InternalMessageInfo

func (m *Silence) GetSilence() string {
	if m != nil {
		return m.Silence
	}
	return ""
}

func (m *Silence) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

type Config struct {
	Silences             []*Silence `protobuf:"bytes,1,rep,name=silences,proto3" json:"silences,omitempty"`
	JobsOfInterest       []string   `protobuf:"bytes,2,rep,name=jobs_of_interest,json=jobsOfInterest,proto3" json:"jobs_of_interest,omitempty"`
	ExternalIP           string     `protobuf:"bytes,3,opt,name=externalIP,proto3" json:"externalIP,omitempty"`
	Issues               []*Issue   `protobuf:"bytes,4,rep,name=issues,proto3" json:"issues,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{2}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetSilences() []*Silence {
	if m != nil {
		return m.Silences
	}
	return nil
}

func (m *Config) GetJobsOfInterest() []string {
	if m != nil {
		return m.JobsOfInterest
	}
	return nil
}

func (m *Config) GetExternalIP() string {
	if m != nil {
		return m.ExternalIP
	}
	return ""
}

func (m *Config) GetIssues() []*Issue {
	if m != nil {
		return m.Issues
	}
	return nil
}

type Issue struct {
	Title                string           `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Body                 string           `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	Service              string           `protobuf:"bytes,3,opt,name=service,proto3" json:"service,omitempty"`
	Number               int32            `protobuf:"varint,4,opt,name=number,proto3" json:"number,omitempty"`
	State                Issue_IssueState `protobuf:"varint,5,opt,name=state,proto3,enum=githubcard.Issue_IssueState" json:"state,omitempty"`
	Sticky               bool             `protobuf:"varint,6,opt,name=sticky,proto3" json:"sticky,omitempty"`
	Origin               Issue_Origin     `protobuf:"varint,7,opt,name=origin,proto3,enum=githubcard.Issue_Origin" json:"origin,omitempty"`
	DateAdded            int64            `protobuf:"varint,8,opt,name=date_added,json=dateAdded,proto3" json:"date_added,omitempty"`
	Url                  string           `protobuf:"bytes,9,opt,name=url,proto3" json:"url,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Issue) Reset()         { *m = Issue{} }
func (m *Issue) String() string { return proto.CompactTextString(m) }
func (*Issue) ProtoMessage()    {}
func (*Issue) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{3}
}

func (m *Issue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Issue.Unmarshal(m, b)
}
func (m *Issue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Issue.Marshal(b, m, deterministic)
}
func (m *Issue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Issue.Merge(m, src)
}
func (m *Issue) XXX_Size() int {
	return xxx_messageInfo_Issue.Size(m)
}
func (m *Issue) XXX_DiscardUnknown() {
	xxx_messageInfo_Issue.DiscardUnknown(m)
}

var xxx_messageInfo_Issue proto.InternalMessageInfo

func (m *Issue) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *Issue) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

func (m *Issue) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *Issue) GetNumber() int32 {
	if m != nil {
		return m.Number
	}
	return 0
}

func (m *Issue) GetState() Issue_IssueState {
	if m != nil {
		return m.State
	}
	return Issue_OPEN
}

func (m *Issue) GetSticky() bool {
	if m != nil {
		return m.Sticky
	}
	return false
}

func (m *Issue) GetOrigin() Issue_Origin {
	if m != nil {
		return m.Origin
	}
	return Issue_UNKNOWN
}

func (m *Issue) GetDateAdded() int64 {
	if m != nil {
		return m.DateAdded
	}
	return 0
}

func (m *Issue) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

type IssueList struct {
	Issues               []*Issue `protobuf:"bytes,1,rep,name=issues,proto3" json:"issues,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IssueList) Reset()         { *m = IssueList{} }
func (m *IssueList) String() string { return proto.CompactTextString(m) }
func (*IssueList) ProtoMessage()    {}
func (*IssueList) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{4}
}

func (m *IssueList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IssueList.Unmarshal(m, b)
}
func (m *IssueList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IssueList.Marshal(b, m, deterministic)
}
func (m *IssueList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IssueList.Merge(m, src)
}
func (m *IssueList) XXX_Size() int {
	return xxx_messageInfo_IssueList.Size(m)
}
func (m *IssueList) XXX_DiscardUnknown() {
	xxx_messageInfo_IssueList.DiscardUnknown(m)
}

var xxx_messageInfo_IssueList proto.InternalMessageInfo

func (m *IssueList) GetIssues() []*Issue {
	if m != nil {
		return m.Issues
	}
	return nil
}

type SilenceRequest struct {
	Silence              string                      `protobuf:"bytes,1,opt,name=silence,proto3" json:"silence,omitempty"`
	Origin               string                      `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin,omitempty"`
	State                SilenceRequest_SilenceState `protobuf:"varint,2,opt,name=state,proto3,enum=githubcard.SilenceRequest_SilenceState" json:"state,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *SilenceRequest) Reset()         { *m = SilenceRequest{} }
func (m *SilenceRequest) String() string { return proto.CompactTextString(m) }
func (*SilenceRequest) ProtoMessage()    {}
func (*SilenceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{5}
}

func (m *SilenceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SilenceRequest.Unmarshal(m, b)
}
func (m *SilenceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SilenceRequest.Marshal(b, m, deterministic)
}
func (m *SilenceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SilenceRequest.Merge(m, src)
}
func (m *SilenceRequest) XXX_Size() int {
	return xxx_messageInfo_SilenceRequest.Size(m)
}
func (m *SilenceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SilenceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SilenceRequest proto.InternalMessageInfo

func (m *SilenceRequest) GetSilence() string {
	if m != nil {
		return m.Silence
	}
	return ""
}

func (m *SilenceRequest) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *SilenceRequest) GetState() SilenceRequest_SilenceState {
	if m != nil {
		return m.State
	}
	return SilenceRequest_UNKNOWN
}

type SilenceResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SilenceResponse) Reset()         { *m = SilenceResponse{} }
func (m *SilenceResponse) String() string { return proto.CompactTextString(m) }
func (*SilenceResponse) ProtoMessage()    {}
func (*SilenceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{6}
}

func (m *SilenceResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SilenceResponse.Unmarshal(m, b)
}
func (m *SilenceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SilenceResponse.Marshal(b, m, deterministic)
}
func (m *SilenceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SilenceResponse.Merge(m, src)
}
func (m *SilenceResponse) XXX_Size() int {
	return xxx_messageInfo_SilenceResponse.Size(m)
}
func (m *SilenceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SilenceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SilenceResponse proto.InternalMessageInfo

type GetAllRequest struct {
	LatestOnly           bool     `protobuf:"varint,1,opt,name=latest_only,json=latestOnly,proto3" json:"latest_only,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetAllRequest) Reset()         { *m = GetAllRequest{} }
func (m *GetAllRequest) String() string { return proto.CompactTextString(m) }
func (*GetAllRequest) ProtoMessage()    {}
func (*GetAllRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{7}
}

func (m *GetAllRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAllRequest.Unmarshal(m, b)
}
func (m *GetAllRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAllRequest.Marshal(b, m, deterministic)
}
func (m *GetAllRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAllRequest.Merge(m, src)
}
func (m *GetAllRequest) XXX_Size() int {
	return xxx_messageInfo_GetAllRequest.Size(m)
}
func (m *GetAllRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAllRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetAllRequest proto.InternalMessageInfo

func (m *GetAllRequest) GetLatestOnly() bool {
	if m != nil {
		return m.LatestOnly
	}
	return false
}

type GetAllResponse struct {
	Issues               []*Issue `protobuf:"bytes,1,rep,name=issues,proto3" json:"issues,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetAllResponse) Reset()         { *m = GetAllResponse{} }
func (m *GetAllResponse) String() string { return proto.CompactTextString(m) }
func (*GetAllResponse) ProtoMessage()    {}
func (*GetAllResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{8}
}

func (m *GetAllResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetAllResponse.Unmarshal(m, b)
}
func (m *GetAllResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetAllResponse.Marshal(b, m, deterministic)
}
func (m *GetAllResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetAllResponse.Merge(m, src)
}
func (m *GetAllResponse) XXX_Size() int {
	return xxx_messageInfo_GetAllResponse.Size(m)
}
func (m *GetAllResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetAllResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetAllResponse proto.InternalMessageInfo

func (m *GetAllResponse) GetIssues() []*Issue {
	if m != nil {
		return m.Issues
	}
	return nil
}

type RegisterRequest struct {
	Job                  string   `protobuf:"bytes,1,opt,name=job,proto3" json:"job,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterRequest) Reset()         { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()    {}
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{9}
}

func (m *RegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterRequest.Unmarshal(m, b)
}
func (m *RegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterRequest.Marshal(b, m, deterministic)
}
func (m *RegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterRequest.Merge(m, src)
}
func (m *RegisterRequest) XXX_Size() int {
	return xxx_messageInfo_RegisterRequest.Size(m)
}
func (m *RegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterRequest proto.InternalMessageInfo

func (m *RegisterRequest) GetJob() string {
	if m != nil {
		return m.Job
	}
	return ""
}

type RegisterResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterResponse) Reset()         { *m = RegisterResponse{} }
func (m *RegisterResponse) String() string { return proto.CompactTextString(m) }
func (*RegisterResponse) ProtoMessage()    {}
func (*RegisterResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{10}
}

func (m *RegisterResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterResponse.Unmarshal(m, b)
}
func (m *RegisterResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterResponse.Marshal(b, m, deterministic)
}
func (m *RegisterResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterResponse.Merge(m, src)
}
func (m *RegisterResponse) XXX_Size() int {
	return xxx_messageInfo_RegisterResponse.Size(m)
}
func (m *RegisterResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterResponse proto.InternalMessageInfo

type DeleteRequest struct {
	Issue                *Issue   `protobuf:"bytes,1,opt,name=issue,proto3" json:"issue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteRequest) Reset()         { *m = DeleteRequest{} }
func (m *DeleteRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteRequest) ProtoMessage()    {}
func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{11}
}

func (m *DeleteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteRequest.Unmarshal(m, b)
}
func (m *DeleteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteRequest.Marshal(b, m, deterministic)
}
func (m *DeleteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteRequest.Merge(m, src)
}
func (m *DeleteRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteRequest.Size(m)
}
func (m *DeleteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteRequest proto.InternalMessageInfo

func (m *DeleteRequest) GetIssue() *Issue {
	if m != nil {
		return m.Issue
	}
	return nil
}

type DeleteResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteResponse) Reset()         { *m = DeleteResponse{} }
func (m *DeleteResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteResponse) ProtoMessage()    {}
func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{12}
}

func (m *DeleteResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteResponse.Unmarshal(m, b)
}
func (m *DeleteResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteResponse.Marshal(b, m, deterministic)
}
func (m *DeleteResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteResponse.Merge(m, src)
}
func (m *DeleteResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteResponse.Size(m)
}
func (m *DeleteResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteResponse proto.InternalMessageInfo

type PullRequest struct {
	Job                  string   `protobuf:"bytes,1,opt,name=job,proto3" json:"job,omitempty"`
	Branch               string   `protobuf:"bytes,2,opt,name=branch,proto3" json:"branch,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PullRequest) Reset()         { *m = PullRequest{} }
func (m *PullRequest) String() string { return proto.CompactTextString(m) }
func (*PullRequest) ProtoMessage()    {}
func (*PullRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{13}
}

func (m *PullRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullRequest.Unmarshal(m, b)
}
func (m *PullRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullRequest.Marshal(b, m, deterministic)
}
func (m *PullRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullRequest.Merge(m, src)
}
func (m *PullRequest) XXX_Size() int {
	return xxx_messageInfo_PullRequest.Size(m)
}
func (m *PullRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PullRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PullRequest proto.InternalMessageInfo

func (m *PullRequest) GetJob() string {
	if m != nil {
		return m.Job
	}
	return ""
}

func (m *PullRequest) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

type PullResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PullResponse) Reset()         { *m = PullResponse{} }
func (m *PullResponse) String() string { return proto.CompactTextString(m) }
func (*PullResponse) ProtoMessage()    {}
func (*PullResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfced67e3377ee11, []int{14}
}

func (m *PullResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullResponse.Unmarshal(m, b)
}
func (m *PullResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullResponse.Marshal(b, m, deterministic)
}
func (m *PullResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullResponse.Merge(m, src)
}
func (m *PullResponse) XXX_Size() int {
	return xxx_messageInfo_PullResponse.Size(m)
}
func (m *PullResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PullResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PullResponse proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("githubcard.Issue_IssueState", Issue_IssueState_name, Issue_IssueState_value)
	proto.RegisterEnum("githubcard.Issue_Origin", Issue_Origin_name, Issue_Origin_value)
	proto.RegisterEnum("githubcard.SilenceRequest_SilenceState", SilenceRequest_SilenceState_name, SilenceRequest_SilenceState_value)
	proto.RegisterType((*Token)(nil), "githubcard.Token")
	proto.RegisterType((*Silence)(nil), "githubcard.Silence")
	proto.RegisterType((*Config)(nil), "githubcard.Config")
	proto.RegisterType((*Issue)(nil), "githubcard.Issue")
	proto.RegisterType((*IssueList)(nil), "githubcard.IssueList")
	proto.RegisterType((*SilenceRequest)(nil), "githubcard.SilenceRequest")
	proto.RegisterType((*SilenceResponse)(nil), "githubcard.SilenceResponse")
	proto.RegisterType((*GetAllRequest)(nil), "githubcard.GetAllRequest")
	proto.RegisterType((*GetAllResponse)(nil), "githubcard.GetAllResponse")
	proto.RegisterType((*RegisterRequest)(nil), "githubcard.RegisterRequest")
	proto.RegisterType((*RegisterResponse)(nil), "githubcard.RegisterResponse")
	proto.RegisterType((*DeleteRequest)(nil), "githubcard.DeleteRequest")
	proto.RegisterType((*DeleteResponse)(nil), "githubcard.DeleteResponse")
	proto.RegisterType((*PullRequest)(nil), "githubcard.PullRequest")
	proto.RegisterType((*PullResponse)(nil), "githubcard.PullResponse")
}

func init() { proto.RegisterFile("githubcard.proto", fileDescriptor_bfced67e3377ee11) }

var fileDescriptor_bfced67e3377ee11 = []byte{
	// 749 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x55, 0xdd, 0x6e, 0xda, 0x4a,
	0x10, 0xc6, 0x18, 0x0c, 0x0c, 0x81, 0x98, 0x3d, 0x47, 0xe7, 0xf8, 0x70, 0x92, 0x16, 0x6d, 0x2f,
	0xe2, 0x5e, 0x34, 0x8d, 0xa8, 0x94, 0x54, 0x8a, 0x7a, 0x41, 0x89, 0x93, 0x92, 0xa6, 0x10, 0x99,
	0xa6, 0xb9, 0x44, 0x36, 0xde, 0x10, 0x27, 0xae, 0x9d, 0xda, 0x4b, 0x55, 0x5e, 0xaa, 0x0f, 0xd0,
	0xc7, 0xe8, 0x83, 0xf4, 0x19, 0x2a, 0xef, 0x0f, 0x98, 0x40, 0xd5, 0xf4, 0x26, 0xda, 0x6f, 0x66,
	0xf6, 0x9b, 0x99, 0xef, 0xdb, 0x18, 0xd0, 0x27, 0x3e, 0xbd, 0x9e, 0xba, 0x63, 0x27, 0xf6, 0x76,
	0xef, 0xe2, 0x88, 0x46, 0x08, 0x16, 0x11, 0xbc, 0x0d, 0xc5, 0xf7, 0xd1, 0x2d, 0x09, 0xd1, 0xdf,
	0x50, 0xa4, 0xe9, 0xc1, 0x50, 0x5a, 0x8a, 0x59, 0xb1, 0x39, 0xc0, 0x87, 0x50, 0x1a, 0xfa, 0x01,
	0x09, 0xc7, 0x04, 0x19, 0x50, 0x4a, 0xf8, 0x51, 0x94, 0x48, 0x88, 0xfe, 0x01, 0x2d, 0x8a, 0xfd,
	0x89, 0x1f, 0x1a, 0x79, 0x96, 0x10, 0x08, 0x7f, 0x55, 0x40, 0xeb, 0x46, 0xe1, 0x95, 0x3f, 0x41,
	0xcf, 0xa1, 0x2c, 0xaa, 0x13, 0x43, 0x69, 0xa9, 0x66, 0xb5, 0xfd, 0xd7, 0x6e, 0x66, 0x2e, 0xd1,
	0xc3, 0x9e, 0x17, 0x21, 0x13, 0xf4, 0x9b, 0xc8, 0x4d, 0x46, 0xd1, 0xd5, 0xc8, 0x0f, 0x29, 0x89,
	0x49, 0x42, 0x8d, 0x7c, 0x4b, 0x35, 0x2b, 0x76, 0x3d, 0x8d, 0x0f, 0xae, 0x7a, 0x22, 0x8a, 0x1e,
	0x01, 0x90, 0x2f, 0x94, 0xc4, 0xa1, 0x13, 0xf4, 0xce, 0x0d, 0x95, 0x4d, 0x90, 0x89, 0xa0, 0xa7,
	0xa0, 0xf9, 0x49, 0x32, 0x25, 0x89, 0x51, 0x60, 0x8d, 0x1b, 0xd9, 0xc6, 0xbd, 0x34, 0x63, 0x8b,
	0x02, 0xfc, 0x23, 0x0f, 0x45, 0x16, 0x61, 0x6a, 0xf8, 0x34, 0x20, 0x73, 0x35, 0x52, 0x80, 0x10,
	0x14, 0xdc, 0xc8, 0x9b, 0x89, 0x35, 0xd9, 0x99, 0xc9, 0x42, 0xe2, 0xcf, 0xfe, 0x98, 0x88, 0xde,
	0x12, 0xa6, 0xb2, 0x84, 0xd3, 0x8f, 0x2e, 0x89, 0x8d, 0x42, 0x4b, 0x31, 0x8b, 0xb6, 0x40, 0xa8,
	0x0d, 0xc5, 0x84, 0x3a, 0x94, 0x18, 0xc5, 0x96, 0x62, 0xd6, 0xdb, 0x5b, 0x2b, 0xf3, 0xf0, 0xbf,
	0xc3, 0xb4, 0xc6, 0xe6, 0xa5, 0x29, 0x57, 0x42, 0xfd, 0xf1, 0xed, 0xcc, 0xd0, 0x5a, 0x8a, 0x59,
	0xb6, 0x05, 0x42, 0x7b, 0x73, 0xe9, 0x4b, 0x8c, 0xcc, 0x58, 0x25, 0x1b, 0xb0, 0xbc, 0x34, 0x05,
	0x6d, 0x03, 0x78, 0x0e, 0x25, 0x23, 0xc7, 0xf3, 0x88, 0x67, 0x94, 0x5b, 0x8a, 0xa9, 0xda, 0x95,
	0x34, 0xd2, 0x49, 0x03, 0x48, 0x07, 0x75, 0x1a, 0x07, 0x46, 0x85, 0xad, 0x92, 0x1e, 0x31, 0x06,
	0x58, 0xcc, 0x83, 0xca, 0x50, 0x18, 0x9c, 0x5b, 0x7d, 0x3d, 0x87, 0x00, 0xb4, 0xee, 0xd9, 0x60,
	0x68, 0x1d, 0xe9, 0x0a, 0xde, 0x07, 0x8d, 0xb7, 0x41, 0x55, 0x28, 0x5d, 0xf4, 0xdf, 0xf6, 0x07,
	0x97, 0x69, 0xc9, 0x06, 0x94, 0x8f, 0xed, 0xc1, 0xbb, 0xd1, 0xa5, 0xf5, 0x5a, 0x57, 0x50, 0x03,
	0x6a, 0x0c, 0xd9, 0x56, 0xd7, 0xea, 0x7d, 0xb0, 0x6c, 0x3d, 0x8f, 0xf7, 0xa1, 0xc2, 0xb8, 0xcf,
	0xfc, 0x84, 0x66, 0x8c, 0x52, 0x7e, 0x67, 0xd4, 0x37, 0x05, 0xea, 0xf2, 0xcd, 0x90, 0x4f, 0xd3,
	0xf4, 0x19, 0x3c, 0xe4, 0x79, 0xaa, 0xd9, 0xe7, 0x89, 0x5e, 0x49, 0x1f, 0xf2, 0x4c, 0xba, 0x9d,
	0x75, 0x0f, 0x92, 0x93, 0x4b, 0x98, 0xb5, 0x04, 0x1f, 0xc0, 0x46, 0x36, 0xbc, 0xbc, 0x79, 0x15,
	0x4a, 0xc3, 0xde, 0x99, 0xd5, 0xef, 0x5a, 0xba, 0x82, 0x6a, 0x50, 0xb9, 0xe8, 0x4b, 0x98, 0xc7,
	0x0d, 0xd8, 0x9c, 0xd3, 0x27, 0x77, 0x51, 0x98, 0x10, 0xbc, 0x07, 0xb5, 0x13, 0x42, 0x3b, 0x41,
	0x20, 0xb7, 0x79, 0x0c, 0xd5, 0xc0, 0xa1, 0x24, 0xa1, 0xa3, 0x28, 0x0c, 0x66, 0x6c, 0xa3, 0xb2,
	0x0d, 0x3c, 0x34, 0x08, 0x83, 0x19, 0x3e, 0x84, 0xba, 0xbc, 0xc1, 0x39, 0xfe, 0x44, 0xbe, 0x27,
	0xb0, 0x69, 0x93, 0x89, 0x9f, 0x50, 0x12, 0xcb, 0x86, 0x3a, 0xa8, 0x37, 0x91, 0x2b, 0xa4, 0x4b,
	0x8f, 0x18, 0x81, 0xbe, 0x28, 0x12, 0x73, 0xbe, 0x84, 0xda, 0x11, 0x09, 0x08, 0x9d, 0xab, 0xbe,
	0x03, 0x45, 0xc6, 0xc9, 0x2e, 0xae, 0xed, 0xc9, 0xf3, 0x58, 0x87, 0xba, 0xbc, 0x29, 0xb8, 0x0e,
	0xa0, 0x7a, 0x3e, 0x5d, 0x6c, 0xbc, 0x32, 0x40, 0xea, 0x9b, 0x1b, 0x3b, 0xe1, 0xf8, 0x5a, 0x7e,
	0x56, 0x38, 0xc2, 0x75, 0xd8, 0xe0, 0x17, 0x39, 0x51, 0xfb, 0xbb, 0x0a, 0xda, 0x09, 0x6b, 0x8b,
	0xda, 0x50, 0xee, 0x78, 0x1e, 0xff, 0x17, 0x5e, 0x9d, 0xa5, 0xb9, 0x1a, 0xc2, 0x39, 0xf4, 0x0c,
	0xd4, 0x13, 0x42, 0x1f, 0x5c, 0xde, 0x01, 0x8d, 0x0b, 0x8f, 0xfe, 0xcb, 0xa6, 0x97, 0xec, 0x6b,
	0x36, 0xd7, 0xa5, 0xc4, 0xde, 0x39, 0x74, 0xb4, 0xf8, 0xa8, 0x36, 0x7f, 0xfd, 0xe8, 0x9a, 0xff,
	0xaf, 0xcd, 0xcd, 0x59, 0x4e, 0xa1, 0x2a, 0xfd, 0x39, 0x8d, 0x5c, 0xb4, 0x54, 0x7d, 0xcf, 0xdd,
	0xe6, 0xd6, 0xfa, 0xe4, 0x9c, 0xeb, 0x18, 0xaa, 0xdc, 0x1d, 0x2e, 0xdd, 0xd2, 0x66, 0x4b, 0x86,
	0x2f, 0x6f, 0x76, 0xcf, 0xd1, 0x1c, 0x7a, 0x03, 0x8d, 0x6e, 0x4c, 0x1c, 0x4a, 0xb2, 0xce, 0xfe,
	0x9b, 0xbd, 0x92, 0x49, 0x34, 0x8d, 0xd5, 0x84, 0x64, 0x72, 0x35, 0xf6, 0x53, 0xf5, 0xe2, 0x67,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x85, 0x43, 0xa1, 0x4b, 0xbe, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GithubClient is the client API for Github service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GithubClient interface {
	AddIssue(ctx context.Context, in *Issue, opts ...grpc.CallOption) (*Issue, error)
	Get(ctx context.Context, in *Issue, opts ...grpc.CallOption) (*Issue, error)
	GetAll(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (*GetAllResponse, error)
	Silence(ctx context.Context, in *SilenceRequest, opts ...grpc.CallOption) (*SilenceResponse, error)
	RegisterJob(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	DeleteIssue(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	CreatePullRequest(ctx context.Context, in *PullRequest, opts ...grpc.CallOption) (*PullResponse, error)
}

type githubClient struct {
	cc *grpc.ClientConn
}

func NewGithubClient(cc *grpc.ClientConn) GithubClient {
	return &githubClient{cc}
}

func (c *githubClient) AddIssue(ctx context.Context, in *Issue, opts ...grpc.CallOption) (*Issue, error) {
	out := new(Issue)
	err := c.cc.Invoke(ctx, "/githubcard.Github/AddIssue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *githubClient) Get(ctx context.Context, in *Issue, opts ...grpc.CallOption) (*Issue, error) {
	out := new(Issue)
	err := c.cc.Invoke(ctx, "/githubcard.Github/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *githubClient) GetAll(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (*GetAllResponse, error) {
	out := new(GetAllResponse)
	err := c.cc.Invoke(ctx, "/githubcard.Github/GetAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *githubClient) Silence(ctx context.Context, in *SilenceRequest, opts ...grpc.CallOption) (*SilenceResponse, error) {
	out := new(SilenceResponse)
	err := c.cc.Invoke(ctx, "/githubcard.Github/Silence", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *githubClient) RegisterJob(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/githubcard.Github/RegisterJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *githubClient) DeleteIssue(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/githubcard.Github/DeleteIssue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *githubClient) CreatePullRequest(ctx context.Context, in *PullRequest, opts ...grpc.CallOption) (*PullResponse, error) {
	out := new(PullResponse)
	err := c.cc.Invoke(ctx, "/githubcard.Github/CreatePullRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GithubServer is the server API for Github service.
type GithubServer interface {
	AddIssue(context.Context, *Issue) (*Issue, error)
	Get(context.Context, *Issue) (*Issue, error)
	GetAll(context.Context, *GetAllRequest) (*GetAllResponse, error)
	Silence(context.Context, *SilenceRequest) (*SilenceResponse, error)
	RegisterJob(context.Context, *RegisterRequest) (*RegisterResponse, error)
	DeleteIssue(context.Context, *DeleteRequest) (*DeleteResponse, error)
	CreatePullRequest(context.Context, *PullRequest) (*PullResponse, error)
}

func RegisterGithubServer(s *grpc.Server, srv GithubServer) {
	s.RegisterService(&_Github_serviceDesc, srv)
}

func _Github_AddIssue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Issue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).AddIssue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/AddIssue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).AddIssue(ctx, req.(*Issue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Github_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Issue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).Get(ctx, req.(*Issue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Github_GetAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).GetAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/GetAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).GetAll(ctx, req.(*GetAllRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Github_Silence_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SilenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).Silence(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/Silence",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).Silence(ctx, req.(*SilenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Github_RegisterJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).RegisterJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/RegisterJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).RegisterJob(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Github_DeleteIssue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).DeleteIssue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/DeleteIssue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).DeleteIssue(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Github_CreatePullRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PullRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubServer).CreatePullRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/githubcard.Github/CreatePullRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubServer).CreatePullRequest(ctx, req.(*PullRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Github_serviceDesc = grpc.ServiceDesc{
	ServiceName: "githubcard.Github",
	HandlerType: (*GithubServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddIssue",
			Handler:    _Github_AddIssue_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Github_Get_Handler,
		},
		{
			MethodName: "GetAll",
			Handler:    _Github_GetAll_Handler,
		},
		{
			MethodName: "Silence",
			Handler:    _Github_Silence_Handler,
		},
		{
			MethodName: "RegisterJob",
			Handler:    _Github_RegisterJob_Handler,
		},
		{
			MethodName: "DeleteIssue",
			Handler:    _Github_DeleteIssue_Handler,
		},
		{
			MethodName: "CreatePullRequest",
			Handler:    _Github_CreatePullRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "githubcard.proto",
}
