package plugins

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/darkdeathoriginal/gogrambot/handler"
)

func init() {
	handler.NewPlugin("kang").
		Description("Steals a sticker and adds it to your userbot pack").
		Category("Userbot").
		Handle(func(message *telegram.NewMessage) error {
			if !message.IsReply() {
				message.Reply("Reply to a sticker or image to kang it!")
				return nil
			}

			replyMsg, err := message.GetReplyMessage()
			if err != nil {
				return err
			}

			me, err := message.Client.GetMe()
			if err != nil {
				return err
			}

			packName := fmt.Sprintf("%spack01" me.Username)
			if me.Username == "" {
				packName = fmt.Sprintf("gogram_pack_%d_by_id", me.ID)
			}
			packTitle := fmt.Sprintf("Gogram Pack %d", me.ID)

			emoji := "🤔"
			args := message.Args()
			if args != "" {
				emoji = strings.Split(args, " ")[0]
			}

			statusMsg, _ := message.Reply("Kanging sticker...")

			// Extract Document from replied message
			msgObj := replyMsg.Message
			if msgObj == nil || msgObj.Media == nil {
				statusMsg.Edit("Please reply to a proper sticker.")
				return nil
			}

			mediaDoc, ok := msgObj.Media.(*telegram.MessageMediaDocument)
			if !ok || mediaDoc.Document == nil {
				statusMsg.Edit("Please reply to a sticker (document).")
				return nil
			}

			docObj, ok := mediaDoc.Document.(*telegram.DocumentObj)
			if !ok {
				statusMsg.Edit("Invalid sticker document.")
				return nil
			}

			stickerItem := &telegram.InputStickerSetItem{
				Document: &telegram.InputDocumentObj{
					ID:            docObj.ID,
					AccessHash:    docObj.AccessHash,
					FileReference: docObj.FileReference,
				},
				Emoji: emoji,
			}

			// Add to set
			_, err = message.Client.StickersAddStickerToSet(
				&telegram.InputStickerSetShortName{ShortName: packName},
				stickerItem,
			)

			if err != nil {
				// Set might not exist, create it
				if strings.Contains(err.Error(), "STICKERSET_INVALID") {
					statusMsg.Edit("Pack not found! Creating new pack...")
					_, err = message.Client.StickersCreateStickerSet(&telegram.StickersCreateStickerSetParams{
						UserID:    &telegram.InputUserSelf{},
						Title:     packTitle,
						ShortName: packName,
						Stickers:  []*telegram.InputStickerSetItem{stickerItem},
					})
					if err != nil {
						statusMsg.Edit(fmt.Sprintf("Failed to create set: %v", err))
						return err
					}
				} else {
					statusMsg.Edit(fmt.Sprintf("Failed to add to set: %v", err))
					return err
				}
			}

			link := fmt.Sprintf("https://t.me/addstickers/%s", packName)
			statusMsg.Edit(fmt.Sprintf("Sticker kanged! [View Pack](%s)", link))

			return nil
		})
}
