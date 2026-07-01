package logic

import (
	"context"
	"fight-game/pb/mail"
	"fight-game/pb/player/asset"
	"fight-game/pb/player/currency"
	"fight-game/pkg/common/utils"
	"fight-game/service/mail/internal/model"
	"fight-game/service/mail/internal/svc"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MailLogic {
	return &MailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendMail 发送邮件
func (l *MailLogic) SendMail(in *mail.SendMailRequest) (*mail.SendMailResponse, error) {
	mailID := utils.GenUUID()
	now := time.Now()

	// 构建附件
	attachments := make([]model.Attachment, 0, len(in.Attachments))
	for _, a := range in.Attachments {
		attachments = append(attachments, model.Attachment{
			Type:      a.Type,
			ID:        a.Id,
			Amount:    a.Amount,
			AssetType: a.AssetType,
		})
	}

	// 计算过期时间
	var expireAt *time.Time
	if in.ExpireAt > 0 {
		t := time.Unix(in.ExpireAt, 0)
		expireAt = &t
	} else if l.svcCtx.Config.Mail.MailExpire > 0 {
		t := now.Add(time.Duration(l.svcCtx.Config.Mail.MailExpire) * time.Second)
		expireAt = &t
	}

	// 写入邮件正文
	mailBody := model.MailBody{
		ID:          primitive.NewObjectID(),
		MailID:      mailID,
		SenderID:    in.SenderId,
		SenderName:  in.SenderName,
		Title:       in.Title,
		Content:     in.Content,
		MailType:    int32(in.MailType),
		Attachments: attachments,
		CreatedAt:   now,
	}

	// 写入邮箱索引（针对单个接收者）
	mailBox := model.MailBox{
		ID:        primitive.NewObjectID(),
		PlayerID:  in.ReceiverId,
		MailID:    mailID,
		Status:    0, // 未读
		CreatedAt: now,
		ExpireAt:  expireAt,
	}

	// 写入 mailBody 和 mailBox（顺序写入，不依赖事务）
	mailBodyColl := l.svcCtx.MongoDB.Collection(model.CollectionMailBody)
	mailBoxColl := l.svcCtx.MongoDB.Collection(model.CollectionMailBox)

	if _, err := mailBodyColl.InsertOne(l.ctx, mailBody); err != nil {
		logx.Errorf("Insert mailBody error: %v", err)
		return nil, err
	}
	if _, err := mailBoxColl.InsertOne(l.ctx, mailBox); err != nil {
		logx.Errorf("Insert mailBox error: %v", err)
		return nil, err
	}

	logx.Infof("Mail sent: mailId=%s, receiver=%s, title=%s", mailID, in.ReceiverId, in.Title)
	return &mail.SendMailResponse{
		MailId: mailID,
	}, nil
}

// GetMailList 获取邮件列表
func (l *MailLogic) GetMailList(in *mail.GetMailListRequest) (*mail.GetMailListResponse, error) {
	page := int64(in.Page)
	pageSize := int64(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	coll := l.svcCtx.MongoDB.Collection(model.CollectionMailBox)

	// 过滤已删除和非过期邮件
	filter := bson.M{
		"player_id": in.PlayerId,
		"status":    bson.M{"$ne": 3}, // 排除已删除
		"$or": []bson.M{
			{"expire_at": nil},
			{"expire_at": bson.M{"$gte": time.Now()}},
		},
	}

	// 查询总数
	total, err := coll.CountDocuments(l.ctx, filter)
	if err != nil {
		logx.Errorf("CountDocuments error: %v", err)
		return nil, err
	}

	// 分页查询
	skip := (page - 1) * pageSize
	cursor, err := coll.Find(l.ctx, filter, options.Find().
		SetSkip(skip).
		SetLimit(pageSize).
		SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		logx.Errorf("Find error: %v", err)
		return nil, err
	}
	defer cursor.Close(l.ctx)

	var boxes []model.MailBox
	if err := cursor.All(l.ctx, &boxes); err != nil {
		logx.Errorf("Cursor.All error: %v", err)
		return nil, err
	}

	// 批量查询邮件正文获取标题和发送者信息
	mailIDs := make([]string, 0, len(boxes))
	for _, b := range boxes {
		mailIDs = append(mailIDs, b.MailID)
	}

	bodyColl := l.svcCtx.MongoDB.Collection(model.CollectionMailBody)
	bodyCursor, err := bodyColl.Find(l.ctx, bson.M{"mail_id": bson.M{"$in": mailIDs}})
	if err != nil {
		logx.Errorf("Find mail bodies error: %v", err)
		return nil, err
	}
	defer bodyCursor.Close(l.ctx)

	var bodies []model.MailBody
	if err := bodyCursor.All(l.ctx, &bodies); err != nil {
		logx.Errorf("Cursor.All bodies error: %v", err)
		return nil, err
	}

	bodyMap := make(map[string]model.MailBody)
	for _, b := range bodies {
		bodyMap[b.MailID] = b
	}

	// 组装响应
	mails := make([]*mail.MailSummary, 0, len(boxes))
	for _, box := range boxes {
		body, ok := bodyMap[box.MailID]
		if !ok {
			continue
		}

		var expireAt int64
		if box.ExpireAt != nil {
			expireAt = box.ExpireAt.Unix()
		}

		mails = append(mails, &mail.MailSummary{
			MailId:         box.MailID,
			SenderId:       body.SenderID,
			SenderName:     body.SenderName,
			Title:          body.Title,
			Status:         mail.MailStatus(box.Status),
			MailType:       mail.MailType(body.MailType),
			HasAttachments: len(body.Attachments) > 0,
			CreatedAt:      box.CreatedAt.Unix(),
			ExpireAt:       expireAt,
		})
	}

	return &mail.GetMailListResponse{
		Mails: mails,
		Total: int32(total),
	}, nil
}

// DeleteMail 删除邮件
func (l *MailLogic) DeleteMail(in *mail.DeleteMailRequest) (*mail.DeleteMailResponse, error) {
	result, err := l.svcCtx.MongoDB.Collection(model.CollectionMailBox).
		UpdateOne(l.ctx,
			bson.M{
				"player_id": in.PlayerId,
				"mail_id":   in.MailId,
				"status":    bson.M{"$ne": 3}, // 未删除
			},
			bson.M{
				"$set": bson.M{
					"status": 3, // 已删除
				},
			},
		)
	if err != nil {
		logx.Errorf("DeleteMail update error: %v", err)
		return &mail.DeleteMailResponse{Success: false}, err
	}

	logx.Infof("Mail deleted: playerId=%s, mailId=%s", in.PlayerId, in.MailId)
	return &mail.DeleteMailResponse{
		Success: result.ModifiedCount > 0,
	}, nil
}

// GetMailDetail 获取邮件详情
func (l *MailLogic) GetMailDetail(in *mail.GetMailDetailRequest) (*mail.GetMailDetailResponse, error) {
	// 查询邮箱索引
	var box model.MailBox
	err := l.svcCtx.MongoDB.Collection(model.CollectionMailBox).
		FindOne(l.ctx, bson.M{"player_id": in.PlayerId, "mail_id": in.MailId}).
		Decode(&box)
	if err != nil {
		logx.Errorf("Find mail box error: %v", err)
		return nil, err
	}

	// 查询邮件正文
	var body model.MailBody
	err = l.svcCtx.MongoDB.Collection(model.CollectionMailBody).
		FindOne(l.ctx, bson.M{"mail_id": in.MailId}).
		Decode(&body)
	if err != nil {
		logx.Errorf("Find mail body error: %v", err)
		return nil, err
	}

	// 组装附件
	attachments := make([]*mail.Attachment, 0, len(body.Attachments))
	for _, a := range body.Attachments {
		attachments = append(attachments, &mail.Attachment{
			Type:      a.Type,
			Id:        a.ID,
			Amount:    a.Amount,
			AssetType: a.AssetType,
		})
	}

	var readAt, claimedAt, expireAt int64
	if box.ReadAt != nil {
		readAt = box.ReadAt.Unix()
	}
	if box.ClaimedAt != nil {
		claimedAt = box.ClaimedAt.Unix()
	}
	if box.ExpireAt != nil {
		expireAt = box.ExpireAt.Unix()
	}

	return &mail.GetMailDetailResponse{
		Mail: &mail.MailDetail{
			MailId:      box.MailID,
			SenderId:    body.SenderID,
			SenderName:  body.SenderName,
			Title:       body.Title,
			Content:     body.Content,
			Status:      mail.MailStatus(box.Status),
			MailType:    mail.MailType(body.MailType),
			Attachments: attachments,
			CreatedAt:   body.CreatedAt.Unix(),
			ReadAt:      readAt,
			ClaimedAt:   claimedAt,
			ExpireAt:    expireAt,
		},
	}, nil
}

// ReadMail 标记邮件已读
func (l *MailLogic) ReadMail(in *mail.ReadMailRequest) (*mail.ReadMailResponse, error) {
	now := time.Now()

	result, err := l.svcCtx.MongoDB.Collection(model.CollectionMailBox).
		UpdateOne(l.ctx,
			bson.M{
				"player_id": in.PlayerId,
				"mail_id":   in.MailId,
				"status":    bson.M{"$lt": 2}, // 未读或已读但未领取
			},
			bson.M{
				"$set": bson.M{
					"status":  1, // 已读
					"read_at": now,
				},
			},
		)
	if err != nil {
		logx.Errorf("ReadMail update error: %v", err)
		return &mail.ReadMailResponse{Success: false}, err
	}

	return &mail.ReadMailResponse{
		Success: result.ModifiedCount > 0,
	}, nil
}

// ClaimAttachment 领取附件
func (l *MailLogic) ClaimAttachment(in *mail.ClaimAttachmentRequest) (*mail.ClaimAttachmentResponse, error) {
	// 查询邮箱索引
	var box model.MailBox
	err := l.svcCtx.MongoDB.Collection(model.CollectionMailBox).
		FindOne(l.ctx, bson.M{"player_id": in.PlayerId, "mail_id": in.MailId}).
		Decode(&box)
	if err != nil {
		logx.Errorf("Find mail box error: %v", err)
		return &mail.ClaimAttachmentResponse{Success: false}, err
	}

	// 已经领取过
	if box.Status == 2 {
		return &mail.ClaimAttachmentResponse{Success: true}, nil
	}

	// 查询邮件正文获取附件
	var body model.MailBody
	err = l.svcCtx.MongoDB.Collection(model.CollectionMailBody).
		FindOne(l.ctx, bson.M{"mail_id": in.MailId}).
		Decode(&body)
	if err != nil {
		logx.Errorf("Find mail body error: %v", err)
		return &mail.ClaimAttachmentResponse{Success: false}, err
	}

	if len(body.Attachments) == 0 {
		return &mail.ClaimAttachmentResponse{Success: false}, nil
	}

	// 调用 Player 服务发放附件
	for _, att := range body.Attachments {
		switch att.Type {
		case "currency":
			// 货币类型附件：调用 ChangeCurrency 增加货币
			currencyType := parseCurrencyType(att.ID)
			_, err := l.svcCtx.PlayerClient.ChangeCurrency(l.ctx, &currency.ChangeCurrencyRequest{
				PlayerId:     in.PlayerId,
				CurrencyType: currencyType,
				ChangeType:   true, // 增加
				Count:        int64(att.Amount),
				Reason:       "mail_attachment:" + in.MailId,
			})
			if err != nil {
				logx.Errorf("ClaimAttachment change currency failed: player=%s, mailId=%s, err=%v",
					in.PlayerId, in.MailId, err)
				return &mail.ClaimAttachmentResponse{Success: false}, err
			}
		case "asset":
			// 资产类型附件：调用 AddAsset 添加资产
			_, err := l.svcCtx.PlayerClient.AddAsset(l.ctx, &asset.AddAssetRequest{
				PlayerId:  in.PlayerId,
				AssetId:   att.ID,
				AssetType: asset.AssetType(att.AssetType),
				Quantity:  att.Amount,
			})
			if err != nil {
				logx.Errorf("ClaimAttachment add asset failed: player=%s, mailId=%s, err=%v",
					in.PlayerId, in.MailId, err)
				return &mail.ClaimAttachmentResponse{Success: false}, err
			}
		default:
			logx.Errorf("ClaimAttachment unknown attachment type: %s", att.Type)
		}
	}

	// 标记已领取
	now := time.Now()
	result, err := l.svcCtx.MongoDB.Collection(model.CollectionMailBox).
		UpdateOne(l.ctx,
			bson.M{
				"player_id": in.PlayerId,
				"mail_id":   in.MailId,
				"status":    bson.M{"$lt": 2},
			},
			bson.M{
				"$set": bson.M{
					"status":     2, // 已领取
					"claimed_at": now,
				},
			},
		)
	if err != nil {
		logx.Errorf("ClaimAttachment update error: %v", err)
		return &mail.ClaimAttachmentResponse{Success: false}, err
	}

	logx.Infof("ClaimAttachment success: player=%s, mailId=%s, attachments=%v",
		in.PlayerId, in.MailId, body.Attachments)
	return &mail.ClaimAttachmentResponse{
		Success: result.ModifiedCount > 0,
	}, nil
}

// parseCurrencyType 将字符串货币ID转换为 proto CurrencyType
func parseCurrencyType(id string) currency.CurrencyType {
	switch id {
	case "gold":
		return currency.CurrencyType_GOLD
	case "point", "diamond":
		return currency.CurrencyType_Point
	default:
		return currency.CurrencyType_None
	}
}
