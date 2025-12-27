// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useIntl} from 'react-intl';

import ProTalkLogoImage from 'images/protalk-logo.png';

type Props = React.HTMLAttributes<HTMLSpanElement> & {
    width?: number;
    height?: number;
};

export default function ProTalkLogo({width, height, ...props}: Props) {
    const {formatMessage} = useIntl();
    return (
        <span {...props}>
            <img
                src={ProTalkLogoImage}
                alt={formatMessage({id: 'generic_icons.protalk', defaultMessage: 'ProTalk Logo'})}
                style={{
                    width: width ? `${width}px` : '100px',
                    height: height ? `${height}px` : 'auto',
                }}
            />
        </span>
    );
}

