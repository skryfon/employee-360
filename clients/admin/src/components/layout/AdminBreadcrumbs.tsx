import { useLocation } from 'react-router-dom'
import { Breadcrumbs, buildBreadcrumbs, type CrumbConfig } from '@employee360/ui'

/** Route pattern -> crumb label. Add new routes (including `:param` detail routes) here. */
export const CRUMBS: CrumbConfig = {
  '/': 'Dashboard',
  '/invitations': 'Invitations',
  '/invitations/new': 'Invite user',
}

export function AdminBreadcrumbs() {
  const { pathname } = useLocation()
  return (
    <Breadcrumbs
      className="order-last -mx-3 w-[calc(100%+1.5rem)] min-w-0 border-t border-slate-200 px-3 py-2 sm:-mx-4 sm:w-[calc(100%+2rem)] sm:px-4 md:order-none md:mx-0 md:w-auto md:flex-1 md:border-t-0 md:p-0"
      items={buildBreadcrumbs(pathname, CRUMBS, ['/dashboard'])}
    />
  )
}
