alter table leads
    add column privacy_notice_version varchar(64),
    add column privacy_consent_at timestamptz;
