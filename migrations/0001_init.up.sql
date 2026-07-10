create extension if not exists pgcrypto;

create table if not exists plans (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    description text not null default '',
    price_cents bigint not null default 0,
    currency text not null default 'USD',
    interval text not null default 'monthly',
    features jsonb not null default '{}',
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists subscriptions (
    id uuid primary key default gen_random_uuid(),
    owner_id uuid not null,
    plan_id uuid not null references plans(id) on delete restrict,
    status text not null default 'active',
    current_period_start timestamptz not null default now(),
    current_period_end timestamptz,
    canceled_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists payments (
    id uuid primary key default gen_random_uuid(),
    owner_id uuid not null,
    subscription_id uuid references subscriptions(id) on delete set null,
    amount_cents bigint not null,
    currency text not null,
    status text not null default 'pending',
    provider text not null default '',
    provider_payment_id text not null default '',
    metadata jsonb not null default '{}',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_plans_is_active on plans(is_active);
create index if not exists idx_plans_interval on plans(interval);
create index if not exists idx_plans_created_at on plans(created_at desc);

create index if not exists idx_subscriptions_owner_id on subscriptions(owner_id);
create index if not exists idx_subscriptions_plan_id on subscriptions(plan_id);
create index if not exists idx_subscriptions_status on subscriptions(status);
create index if not exists idx_subscriptions_created_at on subscriptions(created_at desc);

create index if not exists idx_payments_owner_id on payments(owner_id);
create index if not exists idx_payments_subscription_id on payments(subscription_id);
create index if not exists idx_payments_status on payments(status);
create index if not exists idx_payments_provider on payments(provider);
create index if not exists idx_payments_created_at on payments(created_at desc);

create or replace function set_updated_at()
returns trigger language plpgsql as $$
begin
    new.updated_at = now();
    return new;
end;
$$;

drop trigger if exists trg_plans_updated_at on plans;
create trigger trg_plans_updated_at before update on plans
for each row execute function set_updated_at();

drop trigger if exists trg_subscriptions_updated_at on subscriptions;
create trigger trg_subscriptions_updated_at before update on subscriptions
for each row execute function set_updated_at();

drop trigger if exists trg_payments_updated_at on payments;
create trigger trg_payments_updated_at before update on payments
for each row execute function set_updated_at();
