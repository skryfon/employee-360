export interface NavItem {
  to: string
  label: string
  /** Short glyph shown when the sidebar is collapsed. */
  short: string
}

/** Add new admin sections here; the sidebar renders this list. */
export const NAV_ITEMS: NavItem[] = [{ to: '/invitations', label: 'Invitations', short: 'In' }]
