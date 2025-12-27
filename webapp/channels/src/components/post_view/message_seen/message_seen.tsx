// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {memo, useState, useId, useEffect} from 'react';
import {
    useFloating,
    autoUpdate,
    offset,
    flip,
    shift,
    useHover,
    useInteractions,
    useRole,
    safePolygon,
    FloatingFocusManager,
} from '@floating-ui/react';

import type {ReadReceipt} from '@mattermost/types/posts';
import type {UserProfile} from '@mattermost/types/users';

import {Client4} from 'mattermost-redux/client';

import MessageSeenUsersPopover from './message_seen_users_popover';

import './message_seen.scss';

type Props = {
    currentUserId: UserProfile['id'];
    list?: Array<{user: UserProfile; seenAt: ReadReceipt['seen_at']}>;
    postId: string;
    seenCount: number;
    channelMemberCount?: number;
    actions: {
        getSeenUsersForPost: (postId: string) => void;
    };
};

function MessageSeen({currentUserId, list, channelMemberCount, postId, actions}: Props) {
    const headingId = useId();
    const [open, setOpen] = useState(false);

    // Fetch read receipts on mount
    useEffect(() => {
        actions.getSeenUsersForPost(postId);
    }, [postId, actions]);

    const {x, y, strategy, context, refs: {setReference, setFloating}} = useFloating({
        open,
        onOpenChange: setOpen,
        placement: 'top-start',
        whileElementsMounted: autoUpdate,
        middleware: [
            offset(5),
            flip({fallbackPlacements: ['bottom-start', 'right'], padding: 12}),
            shift({padding: 12}),
        ],
    });

    const {getReferenceProps, getFloatingProps} = useInteractions([
        useHover(context, {
            enabled: list && list.length > 0,
            mouseOnly: true,
            delay: {open: 300, close: 0},
            restMs: 100,
            handleClose: safePolygon({blockPointerEvents: false}),
        }),
        useRole(context),
    ]);

    // Filter out current user
    const othersWhoSeen = list?.filter((item) => item.user.id !== currentUserId) || [];

    if (othersWhoSeen.length === 0) {
        return null;
    }

    // Get up to 3 users for avatar display
    const displayUsers = othersWhoSeen.slice(0, 3);
    const othersCount = othersWhoSeen.length;

    // Check if all members have seen (excluding the post author who is current user)
    // channelMemberCount includes current user, so we compare with (channelMemberCount - 1)
    const allMembersSeen = channelMemberCount && channelMemberCount > 1 && othersCount >= (channelMemberCount - 1);

    return (
        <>
            <span
                ref={setReference}
                className={`MessageSeenIndicator${allMembersSeen ? ' MessageSeenIndicator--all-seen' : ''}`}
                {...getReferenceProps()}
            >
                {allMembersSeen && (
                    <span className='MessageSeenIndicator__checkmark'>
                        {'✓✓'}
                    </span>
                )}
                <span className='MessageSeenIndicator__avatars'>
                    {displayUsers.map((item, index) => (
                        <img
                            key={item.user.id}
                            className='MessageSeenIndicator__avatar'
                            src={Client4.getProfilePictureUrl(item.user.id, item.user.last_picture_update)}
                            alt={item.user.username}
                            style={{zIndex: 3 - index}}
                        />
                    ))}
                </span>
                <span className='MessageSeenIndicator__count'>
                    {othersCount}
                </span>
            </span>
            {open && (
                <FloatingFocusManager
                    context={context}
                    modal={false}
                >
                    <div
                        ref={setFloating}
                        style={{
                            position: strategy,
                            top: y ?? 0,
                            left: x ?? 0,
                            width: 248,
                            zIndex: 999,
                        }}
                        aria-labelledby={headingId}
                        {...getFloatingProps()}
                    >
                        <MessageSeenUsersPopover
                            list={othersWhoSeen}
                        />
                    </div>
                </FloatingFocusManager>
            )}
        </>
    );
}

export default memo(MessageSeen);

