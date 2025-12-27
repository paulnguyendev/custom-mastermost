// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useCallback} from 'react';
import {FormattedMessage} from 'react-intl';
import {useDispatch, useSelector} from 'react-redux';

import type {Channel} from '@mattermost/types/channels';
import type {Post} from '@mattermost/types/posts';

import {Client4} from 'mattermost-redux/client';
import {getDirectTeammate, getMyChannelMemberships, getSortedAllTeamsUnreadChannels} from 'mattermost-redux/selectors/entities/channels';
import {getAllPosts, getPostIdsInChannel} from 'mattermost-redux/selectors/entities/posts';
import {isCollapsedThreadsEnabled} from 'mattermost-redux/selectors/entities/preferences';
import {getTeam} from 'mattermost-redux/selectors/entities/teams';
import {getUser} from 'mattermost-redux/selectors/entities/users';

import {switchToChannel} from 'actions/views/channel';
import {closeRightHandSide} from 'actions/views/rhs';

import NoResultsIndicator from 'components/no_results_indicator/no_results_indicator';
import {NoResultsVariant} from 'components/no_results_indicator/types';
import SearchResultsHeader from 'components/search_results_header';
import Avatar from 'components/widgets/users/avatar';

import Constants from 'utils/constants';

import type {GlobalState} from 'types/store';

import './unread_all_results.scss';

type Props = {
    shrink: () => void;
};

const UnreadAllResults: React.FC<Props> = ({shrink}) => {
    const dispatch = useDispatch();
    const sortedChannels = useSelector((state: GlobalState) => getSortedAllTeamsUnreadChannels(state));
    const myMembers = useSelector((state: GlobalState) => getMyChannelMemberships(state));
    const crtEnabled = useSelector((state: GlobalState) => isCollapsedThreadsEnabled(state));

    const handleChannelClick = useCallback((channel: Channel) => {
        dispatch(switchToChannel(channel));
        dispatch(closeRightHandSide());
    }, [dispatch]);

    return (
        <div className='sidebar-right__body'>
            <SearchResultsHeader>
                <span>
                    <FormattedMessage
                        id='search_header.unreadAll'
                        defaultMessage='Unread Messages'
                    />
                </span>
            </SearchResultsHeader>
            <div className='search-results-container'>
                {sortedChannels.length === 0 ? (
                    <NoResultsIndicator
                        variant={NoResultsVariant.Mentions}
                        titleValues={{}}
                    />
                ) : (
                    <div className='unread-all-results'>
                        {sortedChannels.map((channel) => (
                            <UnreadChannelItem
                                key={channel.id}
                                channel={channel}
                                membership={myMembers[channel.id]}
                                onClick={handleChannelClick}
                                crtEnabled={crtEnabled}
                            />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

type UnreadChannelItemProps = {
    channel: Channel;
    membership: any;
    onClick: (channel: Channel) => void;
    crtEnabled: boolean;
};

const UnreadChannelItem: React.FC<UnreadChannelItemProps> = ({channel, membership, onClick, crtEnabled}) => {
    const team = useSelector((state: GlobalState) => getTeam(state, channel.team_id));
    const mentionCount = crtEnabled ? membership?.mention_count_root : membership?.mention_count;

    // Get avatar info for DM/GM channels
    const teammate = useSelector((state: GlobalState) => getDirectTeammate(state, channel.id));
    const isDM = channel.type === Constants.DM_CHANNEL;
    const isGM = channel.type === Constants.GM_CHANNEL;

    // Get last post
    const postIds = useSelector((state: GlobalState) => getPostIdsInChannel(state, channel.id));
    const allPosts = useSelector((state: GlobalState) => getAllPosts(state));
    const lastPostAuthor = useSelector((state: GlobalState) => {
        if (!postIds || postIds.length === 0) {
            return null;
        }
        const lastPost = allPosts[postIds[0]];
        if (!lastPost) {
            return null;
        }
        return getUser(state, lastPost.user_id);
    });

    const lastPost: Post | null = postIds && postIds.length > 0 ? allPosts[postIds[0]] : null;

    const handleClick = useCallback(() => {
        onClick(channel);
    }, [channel, onClick]);

    // Get avatar URL
    let avatarUrl = '';
    if (isDM && teammate) {
        avatarUrl = Client4.getProfilePictureUrl(teammate.id, teammate.last_picture_update);
    } else if (isGM) {
        // For GM, show first letter count
        avatarUrl = '';
    }

    // Get channel icon
    const renderAvatar = () => {
        if (isDM && teammate) {
            return (
                <Avatar
                    url={avatarUrl}
                    size='sm'
                />
            );
        }
        if (isGM) {
            return (
                <div className='unread-channel-item__gm-icon'>
                    <i className='icon icon-account-multiple-outline'/>
                </div>
            );
        }
        // Public/Private channel
        return (
            <div className='unread-channel-item__channel-icon'>
                <i className={`icon ${channel.type === Constants.PRIVATE_CHANNEL ? 'icon-lock-outline' : 'icon-globe'}`}/>
            </div>
        );
    };

    // Format last message preview
    const getLastMessagePreview = () => {
        if (!lastPost) {
            return null;
        }
        let message = lastPost.message || '';
        if (lastPost.file_ids && lastPost.file_ids.length > 0 && !message) {
            message = '📎 Attachment';
        }
        // Truncate message
        if (message.length > 50) {
            message = message.substring(0, 50) + '...';
        }
        const authorName = lastPostAuthor?.username || '';
        return authorName ? `${authorName}: ${message}` : message;
    };

    return (
        <div
            className='unread-channel-item'
            onClick={handleClick}
            role='button'
            tabIndex={0}
        >
            <div className='unread-channel-item__avatar'>
                {renderAvatar()}
            </div>
            <div className='unread-channel-item__content'>
                <div className='unread-channel-item__header'>
                    <span className='unread-channel-item__name'>
                        {channel.display_name || channel.name}
                    </span>
                    {team && (
                        <span className='unread-channel-item__team'>
                            {team.display_name}
                        </span>
                    )}
                </div>
                {lastPost && (
                    <div className='unread-channel-item__message'>
                        {getLastMessagePreview()}
                    </div>
                )}
            </div>
            {mentionCount > 0 && (
                <span className='badge badge-notify'>
                    {mentionCount}
                </span>
            )}
        </div>
    );
};

export default UnreadAllResults;

