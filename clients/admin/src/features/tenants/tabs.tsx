import type { ComponentType } from 'react'
import type { TenantDetail } from '@employee360/api-client'
import { useOutletContext } from 'react-router-dom'
import { RenameTenantForm } from './components/RenameTenantForm'
import { DomainsManager } from './components/DomainsManager'

export type OrganizationData = TenantDetail

export interface OrganizationTab {
  key: string
  label: string
  /** URL segment under /settings/organization. */
  path: string
  component: ComponentType<{ tenant: OrganizationData }>
}

const GeneralTab = ({ tenant }: { tenant: OrganizationData }) => <RenameTenantForm name={tenant.name ?? ''} />
const DomainsTab = ({ tenant }: { tenant: OrganizationData }) => <DomainsManager domains={tenant.domains ?? []} />

/** Add a future section (Branding, Departments, ...) with one line here; the first entry is the default. */
export const ORGANIZATION_TABS: OrganizationTab[] = [
  { key: 'general', label: 'General', path: 'general', component: GeneralTab },
  { key: 'domains', label: 'Domains', path: 'domains', component: DomainsTab },
]

export const ORGANIZATION_BASE = '/settings/organization'

/** Route element for a tab: reads the tenant loaded once by OrganizationPage. */
export function OrganizationTabRoute({ tab }: { tab: OrganizationTab }) {
  const tenant = useOutletContext<OrganizationData>()
  const Component = tab.component
  return <Component tenant={tenant} />
}
