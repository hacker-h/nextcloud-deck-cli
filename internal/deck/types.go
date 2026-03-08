package deck

type Board struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Color        string    `json:"color"`
	Archived     bool      `json:"archived"`
	ETag         string    `json:"ETag,omitempty"`
	LastModified int64     `json:"lastModified,omitempty"`
	DeletedAt    int64     `json:"deletedAt,omitempty"`
	Stacks       []Stack   `json:"stacks,omitempty"`
	Labels       []Label   `json:"labels,omitempty"`
	ACL          []ACLRule `json:"acl,omitempty"`
	Users        []User    `json:"users,omitempty"`
	Settings     any       `json:"settings,omitempty"`
}

type Stack struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	BoardID      int64  `json:"boardId"`
	Order        int64  `json:"order"`
	DeletedAt    int64  `json:"deletedAt,omitempty"`
	LastModified int64  `json:"lastModified,omitempty"`
	Cards        []Card `json:"cards,omitempty"`
	ETag         string `json:"ETag,omitempty"`
}

type Card struct {
	ID              int64        `json:"id"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	StackID         int64        `json:"stackId"`
	Type            string       `json:"type,omitempty"`
	Order           int64        `json:"order"`
	Archived        bool         `json:"archived"`
	Duedate         *string      `json:"duedate"`
	DeletedAt       int64        `json:"deletedAt,omitempty"`
	LastModified    int64        `json:"lastModified,omitempty"`
	CreatedAt       int64        `json:"createdAt,omitempty"`
	Overdue         int          `json:"overdue,omitempty"`
	ETag            string       `json:"ETag,omitempty"`
	Labels          []Label      `json:"labels,omitempty"`
	AssignedUsers   []Assignment `json:"assignedUsers,omitempty"`
	Attachments     []Attachment `json:"attachments,omitempty"`
	AttachmentCount int          `json:"attachmentCount,omitempty"`
	Owner           any          `json:"owner,omitempty"`
	CommentsUnread  int          `json:"commentsUnread,omitempty"`
}

type Label struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Color   string `json:"color"`
	BoardID int64  `json:"boardId,omitempty"`
	CardID  *int64 `json:"cardId,omitempty"`
}

type Assignment struct {
	ID          int64 `json:"id"`
	CardID      int64 `json:"cardId,omitempty"`
	Participant User  `json:"participant"`
}

type Attachment struct {
	ID           int64          `json:"id"`
	CardID       int64          `json:"cardId"`
	Type         string         `json:"type"`
	Data         string         `json:"data"`
	CreatedAt    int64          `json:"createdAt,omitempty"`
	DeletedAt    int64          `json:"deletedAt,omitempty"`
	CreatedBy    string         `json:"createdBy,omitempty"`
	LastModified int64          `json:"lastModified,omitempty"`
	ExtendedData map[string]any `json:"extendedData,omitempty"`
}

type User struct {
	PrimaryKey  string `json:"primaryKey,omitempty"`
	UID         string `json:"uid,omitempty"`
	DisplayName string `json:"displayname,omitempty"`
}

type ACLRule struct {
	ID               int64 `json:"id"`
	BoardID          int64 `json:"boardId,omitempty"`
	Type             int   `json:"type"`
	Owner            bool  `json:"owner,omitempty"`
	PermissionEdit   bool  `json:"permissionEdit"`
	PermissionShare  bool  `json:"permissionShare"`
	PermissionManage bool  `json:"permissionManage"`
	Participant      User  `json:"participant"`
}

type Comment struct {
	ID               int64     `json:"id"`
	ObjectID         int64     `json:"objectId"`
	Message          string    `json:"message"`
	ActorID          string    `json:"actorId,omitempty"`
	ActorType        string    `json:"actorType,omitempty"`
	ActorDisplayName string    `json:"actorDisplayName,omitempty"`
	CreationDateTime string    `json:"creationDateTime,omitempty"`
	ReplyTo          *Comment  `json:"replyTo,omitempty"`
	Mentions         []Mention `json:"mentions,omitempty"`
}

type Mention struct {
	MentionID          string `json:"mentionId"`
	MentionType        string `json:"mentionType"`
	MentionDisplayName string `json:"mentionDisplayName"`
}

type OCSMeta struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
	Message    string `json:"message"`
}

type OCSResponse[T any] struct {
	OCS struct {
		Meta OCSMeta `json:"meta"`
		Data T       `json:"data"`
	} `json:"ocs"`
}

type BoardCreateRequest struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

type BoardUpdateRequest struct {
	Title    string `json:"title"`
	Color    string `json:"color"`
	Archived bool   `json:"archived"`
}

type CreateACLRuleRequest struct {
	Type             int    `json:"type"`
	Participant      string `json:"participant"`
	PermissionEdit   bool   `json:"permissionEdit"`
	PermissionShare  bool   `json:"permissionShare"`
	PermissionManage bool   `json:"permissionManage"`
}

type UpdateACLRuleRequest struct {
	PermissionEdit   bool `json:"permissionEdit"`
	PermissionShare  bool `json:"permissionShare"`
	PermissionManage bool `json:"permissionManage"`
}

type CreateCardRequest struct {
	Title       string  `json:"title"`
	Type        string  `json:"type,omitempty"`
	Order       int64   `json:"order,omitempty"`
	Description *string `json:"description,omitempty"`
	Duedate     *string `json:"duedate,omitempty"`
}

type UpdateCardRequest struct {
	Title       string  `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type,omitempty"`
	Order       *int64  `json:"order,omitempty"`
	Duedate     *string `json:"duedate"`
	Archived    *bool   `json:"archived,omitempty"`
	Owner       any     `json:"owner,omitempty"`
}

type ReorderCardRequest struct {
	Order   int64 `json:"order"`
	StackID int64 `json:"stackId"`
}

type CreateStackRequest struct {
	Title string `json:"title"`
	Order int64  `json:"order,omitempty"`
}

type UpdateStackRequest struct {
	Title string `json:"title"`
	Order int64  `json:"order"`
}

type CreateLabelRequest struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

type UpdateLabelRequest struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

type AssignLabelRequest struct {
	LabelID int64 `json:"labelId"`
}

type AssignUserRequest struct {
	UserID string `json:"userId"`
}

type CreateCommentRequest struct {
	Message string `json:"message"`
}

type UpdateCommentRequest struct {
	Message string `json:"message"`
}

type ConfigValueRequest struct {
	Value any `json:"value"`
}
