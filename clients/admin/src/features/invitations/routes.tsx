import { Route } from 'react-router-dom'
import InvitationsPage from './pages/InvitationsPage'

/** Authenticated routes; mount inside the RequireAuth + AdminShell layout. */
export const invitationRoutes = <Route path="/invitations" element={<InvitationsPage />} />
