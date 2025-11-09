package services

// User 用户结构体
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetUserByID 根据ID获取用户信息
func GetUserByID(id int) (*User, error) {
	// 这里应该从数据库或其他数据源获取用户信息
	// 现在我们只是返回一个示例用户
	if id == 1 {
		return &User{
			ID:   1,
			Name: "Bilibili User",
		}, nil
	}
	return nil, nil
}