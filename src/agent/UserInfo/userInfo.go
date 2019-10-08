package UserInfo

import "rilihong/RiServer/base"

type UserInfo struct {
	uid int64
	session *base.Session
}

func NewUserInfo(session *base.Session) *UserInfo{
	userInfo := new(UserInfo)
	userInfo.session = session
	return userInfo
}

func (userInfo *UserInfo)SetUserId(uid int64){
	userInfo.uid = uid
}

func (userInfo *UserInfo)GetUserIp() string{
	return userInfo.session.GetConn().RemoteAddr().String()
}

