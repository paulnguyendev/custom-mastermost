// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';
import type {Dispatch} from 'redux';

import type {Post} from '@mattermost/types/posts';

import {getSeenUsersForPost} from 'mattermost-redux/actions/posts';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/common';
import {makeGetReadReceiptsWithProfiles, getReadReceiptCount} from 'mattermost-redux/selectors/entities/posts';

import type {GlobalState} from 'types/store';

import MessageSeen from './message_seen';

type OwnProps = {
    postId: Post['id'];
    channelMemberCount?: number;
};

function makeMapStateToProps() {
    const getReadReceiptsWithProfiles = makeGetReadReceiptsWithProfiles();

    return (state: GlobalState, ownProps: OwnProps) => {
        const currentUserId = getCurrentUserId(state);
        const list = getReadReceiptsWithProfiles(state, ownProps.postId);
        const seenCount = getReadReceiptCount(state, ownProps.postId);

        return {
            currentUserId,
            list,
            seenCount,
            channelMemberCount: ownProps.channelMemberCount,
        };
    };
}

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators({
            getSeenUsersForPost,
        }, dispatch),
    };
}

export default connect(makeMapStateToProps, mapDispatchToProps)(MessageSeen);

