import { Navigate, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { getErrorMessage } from '@employee360/api-client'
import { InlineAlert, PageContainer, PageHeader, Tabs } from '@employee360/ui'
import { OrganizationSkeleton } from '../components/OrganizationSkeleton'
import { useTenantQuery } from '../queries/tenantQueries'
import { ORGANIZATION_BASE, ORGANIZATION_TABS } from '../tabs'

/** Layout for /settings/organization/*: header + URL-driven tabs; the tenant is fetched once and shared via Outlet context. */
export default function OrganizationPage() {
  const { data: tenant, isPending, isError, error } = useTenantQuery()
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const segment = pathname.slice(ORGANIZATION_BASE.length).split('/').filter(Boolean)[0]
  const active = ORGANIZATION_TABS.find((t) => t.path === segment)
  if (!active) return <Navigate to={`${ORGANIZATION_BASE}/${ORGANIZATION_TABS[0].path}`} replace />

  return (
    <PageContainer>
      <PageHeader title="Organization" description="Manage your organization's details and the email domains users can be invited with." />
      <Tabs
        label="Organization sections"
        items={ORGANIZATION_TABS.map((t) => ({ value: t.path, label: t.label }))}
        value={active.path}
        onValueChange={(v) => navigate(`${ORGANIZATION_BASE}/${v}`)}
      >
        <div className="w-full max-w-3xl">
          {isPending && (
            <OrganizationSkeleton tab={active.path} />
          )}
          {isError && <InlineAlert tone="error">{getErrorMessage(error, 'Could not load organization.')}</InlineAlert>}
          {tenant && <Outlet context={tenant} />}
        </div>
      </Tabs>
    </PageContainer>
  )
}
