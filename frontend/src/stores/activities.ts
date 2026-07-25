import { defineStore } from 'pinia'
import { api } from '@/lib/api'

export interface Activity {
  id: string
  description: string
  start_time: string // ISO 8601 (UTC)
  end_time: string // ISO 8601 (UTC)
  type_id: string | null
}

export interface ActivityPayload {
  description: string
  start_time: string
  end_time: string
  type_id: string | null
}

interface ActivitiesState {
  items: Activity[]
}

interface ListResponse {
  activities: Activity[]
}

interface ItemResponse {
  activity: Activity
}

export const useActivitiesStore = defineStore('activities', {
  state: (): ActivitiesState => ({
    items: [],
  }),

  actions: {
    // Load the current user's activities whose start falls in [from, to).
    async fetchRange(from: Date, to: Date): Promise<void> {
      const qs = `from=${encodeURIComponent(from.toISOString())}&to=${encodeURIComponent(
        to.toISOString(),
      )}`
      const { activities } = await api.get<ListResponse>(`/activities?${qs}`)
      this.items = activities
    },

    async create(payload: ActivityPayload): Promise<Activity> {
      const { activity } = await api.post<ItemResponse>('/activities', payload)
      this.items.push(activity)
      return activity
    },

    async update(id: string, payload: ActivityPayload): Promise<Activity> {
      const { activity } = await api.put<ItemResponse>(`/activities/${id}`, payload)
      const i = this.items.findIndex((a) => a.id === id)
      if (i !== -1) this.items[i] = activity
      return activity
    },

    async remove(id: string): Promise<void> {
      await api.delete(`/activities/${id}`)
      this.items = this.items.filter((a) => a.id !== id)
    },
  },
})
