import { defineStore } from 'pinia'
import { api } from '@/lib/api'

export interface DayNote {
  date: string // YYYY-MM-DD
  content: string
}

interface DayNotesState {
  // Notes for the visible range, keyed by YYYY-MM-DD.
  byDate: Record<string, string>
}

interface ListResponse {
  day_notes: DayNote[]
}

interface ItemResponse {
  day_note: DayNote | null
}

export const useDayNotesStore = defineStore('dayNotes', {
  state: (): DayNotesState => ({ byDate: {} }),

  actions: {
    // Load notes for [from, to) (YYYY-MM-DD), replacing the current map.
    async fetchRange(from: string, to: string): Promise<void> {
      const { day_notes } = await api.get<ListResponse>(`/day-notes?from=${from}&to=${to}`)
      this.byDate = Object.fromEntries(day_notes.map((n) => [n.date, n.content]))
    },

    contentFor(date: string): string {
      return this.byDate[date] ?? ''
    },

    // Upsert the note for a day; empty content removes it.
    async save(date: string, content: string): Promise<void> {
      const { day_note } = await api.put<ItemResponse>(`/day-notes/${date}`, { content })
      if (day_note) this.byDate[date] = day_note.content
      else delete this.byDate[date]
    },
  },
})
