export function UserBadge({ email }: { email?: string }) {
  if (!email) return null
  return (
    <span className="flex min-w-0 items-center gap-2" title={email}>
      <span
        aria-hidden="true"
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-sm border border-line-strong bg-surface-subtle text-xs font-semibold uppercase text-ink"
      >
        {email.charAt(0)}
      </span>
      <span className="hidden max-w-[10rem] lg:max-w-[16rem] truncate text-sm text-ink-muted sm:inline">{email}</span>
    </span>
  )
}
