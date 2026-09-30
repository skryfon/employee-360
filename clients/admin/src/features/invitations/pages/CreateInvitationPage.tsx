import { Link, useNavigate } from 'react-router-dom'
import { InviteUserForm } from '../components/InviteUserForm'

export default function CreateInvitationPage() {
  const navigate = useNavigate()
  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-4">
      <Link
        to="/invitations"
        className="text-sm text-slate-900 underline hover:text-slate-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
      >
        Back to invitations
      </Link>
      <h1 className="text-xl font-semibold text-slate-900">Create invitation</h1>
      <InviteUserForm onSuccess={() => navigate('/invitations')} onCancel={() => navigate('/invitations')} />
    </div>
  )
}
