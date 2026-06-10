-- Outbox исходящих Telegram-уведомлений (техдолг #11/#15): уведомление сначала
-- сохраняется здесь, отдельный диспетчер шлёт с ретраями и экспоненциальным backoff —
-- сбой Telegram или рестарт процесса больше не теряет сообщение. Текст отрендерен при
-- постановке, чтобы диспетчер не зависел от доменных связей.
create table notifications (
    id              uuid primary key default gen_random_uuid(),
    kind            varchar(30) not null,
    chat_id         bigint not null,
    text            text not null,
    status          varchar(10) not null default 'pending'
                    check (status in ('pending', 'sent', 'failed')),
    attempts        int not null default 0,
    next_attempt_at timestamptz not null default now(),
    last_error      text,
    sent_at         timestamptz,
    created_at      timestamptz not null default now()
);

-- Частичный индекс под главный запрос диспетчера: pending, отсортированные по сроку.
create index idx_notifications_due on notifications (next_attempt_at) where status = 'pending';
