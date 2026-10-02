import type { ComponentType, SVGProps } from 'react'
import { DashboardIcon, MailIcon } from './NavIcons'

export interface NavItem {
  to: string
  label: string
  /** Icon shown beside the label, and alone when the sidebar is collapsed. */
  icon: ComponentType<SVGProps<SVGSVGElement>>
}

/** Add new admin sections here; the sidebar renders this list. */
export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: DashboardIcon },
  { to: '/invitations', label: 'Invitations', icon: MailIcon },
]
