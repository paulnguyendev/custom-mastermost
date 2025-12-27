-- Drop existing table if exists (to recreate with new schema)
DROP TABLE IF EXISTS ReadReceipts;

CREATE TABLE ReadReceipts (
    PostId VARCHAR(26) NOT NULL,
    UserId VARCHAR(26) NOT NULL,
    ChannelId VARCHAR(26) NOT NULL DEFAULT '',
    SeenAt BIGINT NOT NULL DEFAULT 0,
    ExpireAt BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (PostId, UserId)
);

CREATE INDEX idx_readreceipts_postid ON ReadReceipts(PostId);
CREATE INDEX idx_readreceipts_userid ON ReadReceipts(UserId);
CREATE INDEX idx_readreceipts_channelid ON ReadReceipts(ChannelId);
CREATE INDEX idx_readreceipts_seenat ON ReadReceipts(SeenAt);

