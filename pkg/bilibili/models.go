package bilibili

// CommentMember 评论用户信息
type CommentMember struct {
	Mid       string `json:"mid"`
	Uname     string `json:"uname"`
	Sex       string `json:"sex"`
	Sign      string `json:"sign"`
	Avatar    string `json:"avatar"`
	Rank      string `json:"rank"`
	LevelInfo struct {
		CurrentLevel int `json:"current_level"`
	} `json:"level_info"`
}

// CommentContent 评论内容
type CommentContent struct {
	Message string             `json:"message"`
	Emote   map[string]Emote   `json:"emote"`
	JumpUrl map[string]JumpUrl `json:"jump_url"`
}

// Emote 表情信息
type Emote struct {
	ID        int    `json:"id"`
	PackageID int    `json:"package_id"`
	State     int    `json:"state"`
	Type      int    `json:"type"`
	Attr      int    `json:"attr"`
	Text      string `json:"text"`
	URL       string `json:"url"`
}

// JumpUrl 跳转链接信息
type JumpUrl struct {
	Title string `json:"title"`
	State int    `json:"state"`
}

// CommentData 评论数据结构
type CommentData struct {
	RPID      int64          `json:"rpid"`
	OID       int64          `json:"oid"`
	Type      int            `json:"type"`
	Mid       int64          `json:"mid"`
	Root      int64          `json:"root"`
	Parent    int64          `json:"parent"`
	Dialog    int64          `json:"dialog"`
	Count     int            `json:"count"`
	RCount    int            `json:"rcount"`
	State     int            `json:"state"`
	FansGrade int            `json:"fansgrade"`
	Attr      int            `json:"attr"`
	Ctime     int            `json:"ctime"`
	RpidStr   string         `json:"rpid_str"`
	RootStr   string         `json:"root_str"`
	ParentStr string         `json:"parent_str"`
	Like      int            `json:"like"`
	Action    int            `json:"action"`
	MidStr    string         `json:"mid_str"`
	Content   CommentContent `json:"content"`
	Member    CommentMember  `json:"member"`
}

// CommentResponse 代表评论API的响应
type CommentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Replies []CommentData `json:"replies"`
		Page    struct {
			Count int `json:"count"`
			Num   int `json:"num"`
			Size  int `json:"size"`
		} `json:"page"`
		Cursor struct {
			AllCount        int `json:"all_count"`
			PaginationReply struct {
				NextOffset string `json:"next_offset"`
			} `json:"pagination_reply"`
			Next int `json:"next"`
		} `json:"cursor"`
	} `json:"data"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int64  `json:"mid"`
	Name     string `json:"name"`
	Sex      string `json:"sex"`
	Sign     string `json:"sign"`
	Level    int    `json:"level"`
	Face     string `json:"face"` // 头像URL
	Coins    int    `json:"coins"`
	Birthday string `json:"birthday"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    UserInfo `json:"data"`
}

// VideoInfo 视频信息
type VideoInfo struct {
	BVID      string `json:"bvid"`
	AID       int64  `json:"aid"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Created   int64  `json:"created"`
	Duration  int    `json:"duration"`
	Pic       string `json:"pic"`
	Owner     Owner  `json:"owner"`
	Stat      Stat   `json:"stat"`
	Copyright int    `json:"copyright"`
}

// Owner 视频所有者信息
type Owner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

// Stat 视频统计数据
type Stat struct {
	Aid       int64  `json:"aid"`
	View      int    `json:"view"`
	Danmaku   int    `json:"danmaku"`
	Reply     int    `json:"reply"`
	Favorite  int    `json:"favorite"`
	Coin      int    `json:"coin"`
	Share     int    `json:"share"`
	Like      int    `json:"like"`
	Dislike   int    `json:"dislike"`
	NowRank   int    `json:"now_rank"`
	HisRank   int    `json:"his_rank"`
	NoReprint int    `json:"no_reprint"`
	Copyright int    `json:"copyright"`
	ArgueMsg  string `json:"argue_msg"`
}

// VideoResponse 视频信息响应
type VideoResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    VideoInfo `json:"data"`
}
