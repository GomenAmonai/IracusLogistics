alter table leads
    drop column if exists privacy_consent_at,
    drop column if exists privacy_notice_version;
