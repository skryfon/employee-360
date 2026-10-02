import type { ComponentType, SVGProps } from 'react'
import { BuildingIcon, DashboardIcon, MailIcon } from './NavIcons'

export interface NavItem {
  to: string
  label: string
  /** Icon shown beside the label, and alone when the sidebar is collapsed. */
  icon: ComponentType<SVGProps<SVGSVGElement>>
  /** When set, the item is shown only to users holding one of these roles. */
  roles?: string[]
}

/** Add new admin sections here; the sidebar renders this list. */
export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: DashboardIcon },
  { to: '/invitations', label: 'Invitations', icon: MailIcon },
  { to: '/settings/organization', label: 'Organization', icon: BuildingIcon, roles: ['super_admin'] },
]

export function visibleNavItems(roles: string[] | undefined): NavItem[] {
  return NAV_ITEMS.filter((i) => !i.roles || i.roles.some((r) => roles?.includes(r)))
}
