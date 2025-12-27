// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {FormattedMessage, useIntl} from 'react-intl';
import {useDispatch, useSelector} from 'react-redux';

import {getAllTeamsUnreadCount} from 'mattermost-redux/selectors/entities/channels';

import {closeRightHandSide, showUnreadAll} from 'actions/views/rhs';
import {getRhsState} from 'selectors/rhs';

import IconButton from 'components/global_header/header_icon_button';
import WithTooltip from 'components/with_tooltip';

import {RHSStates} from 'utils/constants';

import type {GlobalState} from 'types/store';

const UnreadAllButton = (): JSX.Element => {
    const {formatMessage} = useIntl();
    const dispatch = useDispatch();
    const rhsState = useSelector((state: GlobalState) => getRhsState(state));
    const unreadCount = useSelector((state: GlobalState) => getAllTeamsUnreadCount(state));

    const unreadAllButtonClick = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        if (rhsState === RHSStates.UNREAD_ALL) {
            dispatch(closeRightHandSide());
        } else {
            dispatch(showUnreadAll());
        }
    };

    return (
        <WithTooltip
            title={
                <FormattedMessage
                    id='channel_header.unreadAll'
                    defaultMessage='Unread messages'
                />
            }
        >
            <div style={{position: 'relative', display: 'inline-flex'}}>
                <IconButton
                    icon={'email-outline'}
                    toggled={rhsState === RHSStates.UNREAD_ALL}
                    onClick={unreadAllButtonClick}
                    aria-expanded={rhsState === RHSStates.UNREAD_ALL}
                    aria-controls='searchContainer'
                    aria-label={formatMessage({id: 'channel_header.unreadAll', defaultMessage: 'Unread messages'})}
                />
                {unreadCount > 0 && (
                    <span
                        className='badge badge-notify'
                        style={{
                            position: 'absolute',
                            top: '-4px',
                            right: '-4px',
                            fontSize: '10px',
                            minWidth: '16px',
                            height: '16px',
                            lineHeight: '16px',
                            padding: '0 4px',
                        }}
                    >
                        {unreadCount > 99 ? '99+' : unreadCount}
                    </span>
                )}
            </div>
        </WithTooltip>
    );
};

export default UnreadAllButton;

