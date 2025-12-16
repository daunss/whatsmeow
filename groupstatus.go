// Copyright (c) 2025 Daunss (fork)
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// SendStatusToGroup sends a status only visible to members of a specific group.
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
