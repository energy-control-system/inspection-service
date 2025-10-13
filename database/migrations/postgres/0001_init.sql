-- +goose Up
create table if not exists inspection_statuses
(
    id   int primary key generated always as identity,
    name text not null
);

insert into inspection_statuses (name)
values ('InWork'),
       ('Done');

create table if not exists inspection_types
(
    id   int primary key generated always as identity,
    name text not null
);

insert into inspection_types (name)
values ('Limitation'),
       ('Resumption'),
       ('Verification'),
       ('UnauthorizedConnection');

create table if not exists inspection_resolutions
(
    id   int primary key generated always as identity,
    name text not null
);

insert into inspection_resolutions (name)
values ('Limited'),
       ('Stopped'),
       ('Resumed');

create table if not exists inspection_methods_by
(
    id   int primary key generated always as identity,
    name text not null
);

insert into inspection_methods_by (name)
values ('Consumer'),
       ('Inspector');

create table if not exists inspection_reason_types
(
    id   int primary key generated always as identity,
    name text not null
);

insert into inspection_reason_types (name)
values ('NotIntroduced'),
       ('ConsumerLimited'),
       ('InspectorLimited'),
       ('Resumed');

create table if not exists inspections
(
    id                        int primary key generated always as identity,
    task_id                   int         not null unique,
    status                    int         not null references inspection_statuses (id) on delete restrict,
    type                      int references inspection_types (id) on delete restrict,
    resolution                int references inspection_resolutions (id) on delete restrict, -- Результаты проверки
    limit_reason              text,                                                          -- Основание введения ограничения (приостановления) режима потребления. Если NULL, то это неполная оплата
    method                    text,                                                          -- Способ введения ограничения, приостановления, возобновления режима потребления, номера и место установки пломб
    method_by                 int references inspection_methods_by (id) on delete restrict,  -- Кем было введено
    reason_type               int references inspection_reason_types (id) on delete restrict,
    reason_description        text,                                                          -- Причина невведения ограничения (только для reason_type = 1)
    is_restriction_checked    bool,                                                          -- Произведена проверка введенного ограничения
    is_violation_detected     bool,                                                          -- Нарушение потребителем введенного ограничения выявлено
    is_expense_available      bool,                                                          -- Наличие расхода после введенного ограничения
    violation_description     text,                                                          -- Иное описание выявленного нарушения/сведения, на основании которых сделан вывод о нарушении
    is_unauthorized_consumers bool,                                                          -- Самовольное подключение энергопринимающих устройств Потребителя к электрическим сетям
    unauthorized_description  text,                                                          -- Описание места и способа самовольного подключения к электрическим сетям
    unauthorized_explanation  text,                                                          -- Объяснение лица, допустившего самовольное подключение к электрическим сетям
    inspect_at                timestamptz,                                                   -- Дата проверки
    energy_action_at          timestamptz,                                                   -- Время действия над подачей электроэнергии
    created_at                timestamptz not null default now(),
    updated_at                timestamptz not null default now()
);

create table if not exists inspected_devices
(
    id            int primary key generated always as identity,
    device_id     int            not null,
    inspection_id int            not null references inspections (id) on delete cascade,
    value         numeric(15, 2) not null, -- Текущее показание
    consumption   numeric(15, 2) not null, -- Расход электрической энергии кВтч
    created_at    timestamptz    not null default now()
);

create table if not exists inspected_seals
(
    id            int primary key generated always as identity,
    seal_id       int         not null,
    inspection_id int         not null references inspections (id) on delete cascade,
    is_broken     bool        not null, -- Сорвана ли пломба
    created_at    timestamptz not null default now()
);

create table if not exists attachment_types
(
    id   int primary key generated always as identity,
    name text not null
);

insert into attachment_types (name)
values ('DevicePhoto'),
       ('SealPhoto'),
       ('Act');

create table if not exists attachments
(
    id            int primary key generated always as identity,
    inspection_id int         not null references inspections (id) on delete cascade,
    type          int         not null references attachment_types (id) on delete restrict,
    file_id       int         not null,
    created_at    timestamptz not null default now()
);

create index if not exists idx_inspections_task on inspections (task_id);
create index if not exists idx_inspections_status on inspections (status);
create index if not exists idx_inspections_inspect_at on inspections (inspect_at);
create index if not exists idx_devices_inspection on inspected_devices (inspection_id);
create index if not exists idx_seals_inspection on inspected_seals (inspection_id);
create index if not exists idx_attachments_inspection on attachments (inspection_id);
create index if not exists idx_attachments_type on attachments (type);

-- +goose StatementBegin
create or replace function update_updated_at_column()
    returns trigger as
$$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;
-- +goose StatementEnd

create trigger trg_inspections_updated_at
    before update
    on inspections
    for each row
execute function update_updated_at_column();

-- +goose Down
drop trigger if exists trg_inspections_updated_at on inspections;
drop function if exists update_updated_at_column();
drop table if exists attachments;
drop table if exists attachment_types;
drop table if exists inspected_seals;
drop table if exists inspected_devices;
drop table if exists inspections;
drop table if exists inspection_reason_types;
drop table if exists inspection_methods_by;
drop table if exists inspection_resolutions;
drop table if exists inspection_types;
drop table if exists inspection_statuses;
