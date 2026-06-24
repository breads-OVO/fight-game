package utils

import (
	"fight-game/pb/common"

	"google.golang.org/protobuf/proto"
)

// ============================================================
// 构建 helper
// ============================================================

// BuildWSMessage 构建 WSMessage（不执行 Marshal）
func BuildWSMessage(msgType int32, body []byte, playerId string, seq string, frameNo int32) *common.WSMessage {
	if seq == "" {
		seq = GenUUIDWithPrefix("seq-")
	}
	return &common.WSMessage{
		MsgType:   msgType,
		MsgID:     GenUUID(),
		SeqId:     seq,
		Body:      body,
		PlayerId:  playerId,
		Timestamp: GetTimestamp(),
		FrameNo:   frameNo,
	}
}

// BuildWSResponseBody 构建 WSResponse 并 marshal 为 bytes（不依赖请求消息，灵活使用）
func BuildWSResponseBody(code int32, message string, body []byte) ([]byte, error) {
	resp := &common.WSResponse{
		Code:    code,
		Message: message,
		Data:    body,
	}
	return proto.Marshal(resp)
}

// BuildWSResponse 根据请求消息构建响应 WSMessage（自动关联 seqId + playerId）
func BuildWSResponse(code int32, message string, body []byte, wsMsg *common.WSMessage) (*common.WSMessage, error) {
	respBody, err := BuildWSResponseBody(code, message, body)
	if err != nil {
		return nil, err
	}
	return BuildWSMessage(wsMsg.MsgType, respBody, wsMsg.PlayerId, wsMsg.SeqId, 0), nil
}

// ============================================================
// 打包/解包 helper（Marshal / Unmarshal 封装）
// ============================================================

// PackWSMessage 构建 WSMessage 并 marshal 为 []byte（用于发送）
func PackWSMessage(msgType int32, body []byte, playerId string, seq string, frameNo int32) ([]byte, error) {
	msg := BuildWSMessage(msgType, body, playerId, seq, frameNo)
	return proto.Marshal(msg)
}

// PackWSMessageWithProto 构建 WSMessage（body 为 proto.Message）并 marshal 为 []byte
func PackWSMessageWithProto(msgType int32, playerId string, body proto.Message, seq string, frameNo int32) ([]byte, error) {
	bodyBytes, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}
	return PackWSMessage(msgType, bodyBytes, playerId, seq, frameNo)
}

// UnpackWSMessage 从 []byte 反序列化 WSMessage
func UnpackWSMessage(data []byte) (*common.WSMessage, error) {
	var msg common.WSMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// UnpackBody 将 WSMessage.Body 反序列化到目标 proto.Message
func UnpackBody(msg *common.WSMessage, target proto.Message) error {
	return proto.Unmarshal(msg.Body, target)
}
