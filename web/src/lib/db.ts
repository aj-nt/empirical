import { openDB, type IDBPDatabase } from 'idb';
import type { SavedChart, BirthData } from './types';

const DB_NAME = 'empirical';
const DB_VERSION = 1;

let dbPromise: Promise<IDBPDatabase> | null = null;

function getDB(): Promise<IDBPDatabase> {
  if (!dbPromise) {
    dbPromise = openDB(DB_NAME, DB_VERSION, {
      upgrade(db) {
        const store = db.createObjectStore('charts', {
          keyPath: 'id',
          autoIncrement: true,
        });
        store.createIndex('name', 'name');
        store.createIndex('tags', 'tags', { multiEntry: true });
        store.createIndex('updatedAt', 'updatedAt');
      },
    });
  }
  return dbPromise;
}

export const chartDB = {
  async getAll(): Promise<SavedChart[]> {
    const db = await getDB();
    return db.getAll('charts');
  },

  async getById(id: number): Promise<SavedChart | undefined> {
    const db = await getDB();
    return db.get('charts', id);
  },

  async add(chart: Omit<SavedChart, 'id'>): Promise<number> {
    const db = await getDB();
    return db.add('charts', chart) as Promise<number>;
  },

  async update(id: number, chart: Partial<SavedChart>): Promise<void> {
    const db = await getDB();
    const existing = await db.get('charts', id);
    if (!existing) throw new Error(`Chart ${id} not found`);
    const updated = { ...existing, ...chart, updatedAt: new Date().toISOString() };
    await db.put('charts', updated);
  },

  async remove(id: number): Promise<void> {
    const db = await getDB();
    await db.delete('charts', id);
  },

  async search(query: string): Promise<SavedChart[]> {
    const db = await getDB();
    const all = await db.getAll('charts');
    const q = query.toLowerCase();
    return all.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.tags.some((t: string) => t.toLowerCase().includes(q)) ||
        c.notes.toLowerCase().includes(q)
    );
  },

  async getByTag(tag: string): Promise<SavedChart[]> {
    const db = await getDB();
    return db.getAllFromIndex('charts', 'tags', tag);
  },

  async getRecent(limit = 10): Promise<SavedChart[]> {
    const db = await getDB();
    const all = await db.getAll('charts');
    return all
      .sort((a: SavedChart, b: SavedChart) => b.updatedAt.localeCompare(a.updatedAt))
      .slice(0, limit);
  },

  async exportAll(): Promise<string> {
    const db = await getDB();
    const all = await db.getAll('charts');
    return JSON.stringify(all, null, 2);
  },

  async importAll(json: string): Promise<number> {
    const charts: SavedChart[] = JSON.parse(json);
    const db = await getDB();
    const tx = db.transaction('charts', 'readwrite');
    let count = 0;
    for (const chart of charts) {
      // Remove id to let autoIncrement assign new ones
      const { id, ...rest } = chart;
      await tx.store.add({ ...rest, createdAt: rest.createdAt || new Date().toISOString(), updatedAt: new Date().toISOString() });
      count++;
    }
    await tx.done;
    return count;
  },

  async createFromBirthData(name: string, birthData: BirthData, tags: string[] = [], houseSystem = 'placidus'): Promise<number> {
    const now = new Date().toISOString();
    return this.add({
      name,
      birthData,
      houseSystem,
      tags,
      notes: '',
      createdAt: now,
      updatedAt: now,
    });
  },
};
