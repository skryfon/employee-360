import { Navigate, Route } from 'react-router-dom'
import OrganizationPage from './pages/OrganizationPage'
import { RequireSuperAdmin } from './components/RequireSuperAdmin'
import { ORGANIZATION_TABS, OrganizationTabRoute } from './tabs'

/** Authenticated, super_admin-only routes; mount inside the RequireAuth + AdminShell layout. */
export const tenantRoutes = (
  <Route element={<RequireSuperAdmin />}>
    <Route path="/settings/organization" element={<OrganizationPage />}>
      <Route index element={<Navigate to={ORGANIZATION_TABS[0].path} replace />} />
      {ORGANIZATION_TABS.map((tab) => (
        <Route key={tab.key} path={tab.path} element={<OrganizationTabRoute tab={tab} />} />
      ))}
    </Route>
  </Route>
)
