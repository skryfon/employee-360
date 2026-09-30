import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  inviteUser,
  listInvitations,
  resendInvitation,
  revokeInvitation,
  type InviteUserRequest,
} from '@employee360/api-client'

export const INVITATIONS_KEY = ['invitations'] as const
export const INVITATIONS_PAGE_SIZE = 20

export function useInvitationsQuery(page: number) {
  return useQuery({
    queryKey: [...INVITATIONS_KEY, { page }],
    queryFn: ({ signal }) => listInvitations({ page, page_size: INVITATIONS_PAGE_SIZE }, signal),
    placeholderData: keepPreviousData,
  })
}

function useInvalidatingMutation<TVars, TResult>(fn: (vars: TVars) => Promise<TResult>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => qc.invalidateQueries({ queryKey: INVITATIONS_KEY }),
  })
}

export const useInviteUserMutation = () =>
  useInvalidatingMutation((body: InviteUserRequest) => inviteUser(body))
export const useResendInvitationMutation = () =>
  useInvalidatingMutation((id: string) => resendInvitation(id))
export const useRevokeInvitationMutation = () =>
  useInvalidatingMutation((id: string) => revokeInvitation(id))
