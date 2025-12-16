// Copyright (c) 2025 Daunss (fork)
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"context"
	"fmt"

	"go.mau.fi/util/random"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// SendGroupStatus sends a status update that is only visible to members of a specific group.
// This uses the native WhatsApp groupStatusMessageV2 protocol, which allows group members
// to see the status without needing to save the sender's phone number.
//
// Parameters:
//   - groupJID: The JID of the target group (must end with @g.us)
//   - innerMessage: The actual content message (text, image, video, etc.)
//
// Example usage:
//
//	resp, err := cli.SendGroupStatus(ctx, groupJID, &waE2E.Message{
//	    Conversation: proto.String("Hello group members!"),
//	})
//
//	// Or with an image:
//	resp, err := cli.SendGroupStatus(ctx, groupJID, &waE2E.Message{
//	    ImageMessage: imageMsg,
//	})
func (cli *Client) SendGroupStatus(ctx context.Context, groupJID types.JID, innerMessage *waE2E.Message, extra ...SendRequestExtra) (SendResponse, error) {
	if groupJID.Server != types.GroupServer {
		return SendResponse{}, fmt.Errorf("requires group JID, got %s", groupJID.Server)
	}

	// Generate 32-byte random messageSecret (required for group status protocol)
	messageSecret := random.Bytes(32)

	// Wrap the inner message with MessageContextInfo containing messageSecret
	if innerMessage.MessageContextInfo == nil {
		innerMessage.MessageContextInfo = &waE2E.MessageContextInfo{}
	}
	innerMessage.MessageContextInfo.MessageSecret = messageSecret

	// Create the groupStatusMessageV2 wrapper (native protocol)
	wrappedMessage := &waE2E.Message{
		MessageContextInfo: &waE2E.MessageContextInfo{
			MessageSecret: messageSecret,
		},
		GroupStatusMessageV2: &waE2E.FutureProofMessage{
			Message: innerMessage,
		},
	}

	// Send directly to the group JID (not status@broadcast)
	return cli.SendMessage(ctx, groupJID, wrappedMessage, extra...)
}

// SendStatusToGroup sends a status only visible to members of a specific group.
// Deprecated: Use SendGroupStatus instead for native group status protocol support.
// This function uses the legacy approach of sending to status@broadcast with StatusRecipients.
func (cli *Client) SendStatusToGroup(ctx context.Context, groupJID types.JID, message *waE2E.Message, extra ...SendRequestExtra) (SendResponse, error) {
	if groupJID.Server != types.GroupServer {
		return SendResponse{}, fmt.Errorf("requires group JID, got %s", groupJID.Server)
	}

	groupInfo, err := cli.GetGroupInfo(ctx, groupJID)
	if err != nil {
		return SendResponse{}, fmt.Errorf("failed to get group info: %w", err)
	}

	participants := make([]types.JID, 0, len(groupInfo.Participants))
	for _, p := range groupInfo.Participants {
		participants = append(participants, p.JID)
	}

	if len(participants) == 0 {
		return SendResponse{}, fmt.Errorf("group has no participants")
	}

	var req SendRequestExtra
	if len(extra) > 0 {
		req = extra[0]
	}
	req.StatusRecipients = participants

	return cli.SendMessage(ctx, types.StatusBroadcastJID, message, req)
}

// GetGroupMemberJIDs returns JIDs of all group members.
func (cli *Client) GetGroupMemberJIDs(ctx context.Context, groupJID types.JID) ([]types.JID, error) {
	if groupJID.Server != types.GroupServer {
		return nil, fmt.Errorf("requires group JID")
	}

	groupInfo, err := cli.GetGroupInfo(ctx, groupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	jids := make([]types.JID, len(groupInfo.Participants))
	for i, p := range groupInfo.Participants {
		jids[i] = p.JID
	}
	return jids, nil
}
