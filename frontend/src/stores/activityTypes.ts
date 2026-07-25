import { defineStore } from 'pinia'
import { api } from '@/lib/api'

export interface ActivityType {
  id: string
  name: string
  color: string
  sort: number
}

export interface ActivityTypePayload {
  name: string
  color: string
  sort: number
}

interface ActivityTypesState {
  items: ActivityType[]
}

interface ListResponse {
  activity_types: ActivityType[]
}

interface ItemResponse {
  activity_type: ActivityType
}

export const useActivityTypesStore = defineStore('activityTypes', {
  state: (): ActivityTypesState => ({
    items: [],
  }),

  actions: {
    // Load the current user's activity types (already sorted ascending by the API).
    async fetchAll(): Promise<void> {
      const { activity_types } = await api.get<ListResponse>('/activity-types')
      this.items = activity_types
    },

    async fetchOne(id: string): Promise<ActivityType> {
      const { activity_type } = await api.get<ItemResponse>(`/activity-types/${id}`)
      return activity_type
    },

    async create(payload: ActivityTypePayload): Promise<ActivityType> {
      const { activity_type } = await api.post<ItemResponse>('/activity-types', payload)
      return activity_type
    },

    async update(id: string, payload: ActivityTypePayload): Promise<ActivityType> {
      const { activity_type } = await api.put<ItemResponse>(`/activity-types/${id}`, payload)
      return activity_type
    },

    async remove(id: string): Promise<void> {
      await api.delete(`/activity-types/${id}`)
      this.items = this.items.filter((item) => item.id !== id)
    },
  },
})
