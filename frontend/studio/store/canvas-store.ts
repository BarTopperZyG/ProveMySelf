import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { devtools } from 'zustand/middleware';
import { CanvasItem, CanvasState, EditorMode, Position, Size, ItemType, GridConfig, PreviewConfig } from '../types/quiz';
import { generateId } from '../lib/utils';

interface CanvasStore extends CanvasState {
  // Editor state
  editorMode: EditorMode;
  previewConfig: PreviewConfig;
  gridConfig: GridConfig;
  
  // Actions
  addItem: (type: ItemType, position: Position) => void;
  updateItem: (id: string, updates: Partial<CanvasItem>) => void;
  deleteItem: (id: string) => void;
  deleteSelectedItems: () => void;
  selectItem: (id: string, multiSelect?: boolean) => void;
  deselectAll: () => void;
  moveItem: (id: string, position: Position) => void;
  resizeItem: (id: string, size: Size) => void;
  duplicateItem: (id: string) => void;
  
  // Bulk operations
  copySelectedItems: () => void;
  pasteItems: (position?: Position) => void;
  cutSelectedItems: () => void;
  
  // History management
  undo: () => void;
  redo: () => void;
  canUndo: () => boolean;
  canRedo: () => boolean;
  saveState: () => void;
  
  // Canvas operations
  setGridConfig: (config: Partial<GridConfig>) => void;
  setZoom: (zoom: number) => void;
  setCanvasSize: (size: Size) => void;
  resetCanvas: () => void;
  
  // Editor mode
  setEditorMode: (mode: EditorMode) => void;
  setPreviewConfig: (config: Partial<PreviewConfig>) => void;
  
  // Import/Export
  loadCanvas: (state: CanvasState) => void;
  exportCanvas: () => CanvasState;
}

const DEFAULT_GRID_CONFIG: GridConfig = {
  size: 20,
  snap: true,
  visible: true,
  color: '#e5e7eb'
};

const DEFAULT_PREVIEW_CONFIG: PreviewConfig = {
  mode: 'desktop',
  showNavigation: true,
  autoAdvance: false,
  allowBack: true
};

const createInitialState = (): CanvasState => ({
  items: [],
  selectedItems: [],
  clipboardItems: [],
  history: [],
  historyIndex: -1,
  gridSize: DEFAULT_GRID_CONFIG.size,
  snapToGrid: DEFAULT_GRID_CONFIG.snap,
  showGrid: DEFAULT_GRID_CONFIG.visible,
  zoom: 1,
  canvasSize: { width: 1200, height: 800 }
});

const DEFAULT_ITEM_SIZE: Size = { width: 300, height: 100 };

const ITEM_SIZES: Record<ItemType, Size> = {
  [ItemType.TITLE]: { width: 400, height: 60 },
  [ItemType.MEDIA]: { width: 300, height: 200 },
  [ItemType.CHOICE]: { width: 350, height: 150 },
  [ItemType.MULTI_CHOICE]: { width: 350, height: 180 },
  [ItemType.TEXT_ENTRY]: { width: 300, height: 120 },
  [ItemType.ORDERING]: { width: 300, height: 200 },
  [ItemType.HOTSPOT]: { width: 400, height: 300 }
};

export const useCanvasStore = create<CanvasStore>()(
  devtools(
    immer((set, get) => ({
      ...createInitialState(),
      editorMode: EditorMode.EDIT,
      previewConfig: DEFAULT_PREVIEW_CONFIG,
      gridConfig: DEFAULT_GRID_CONFIG,

      addItem: (type: ItemType, position: Position) => {
        set((state) => {
          const newItem: CanvasItem = {
            id: generateId('item'),
            type,
            title: `New ${type.replace('_', ' ')}`,
            position: state.snapToGrid 
              ? { 
                  x: Math.round(position.x / state.gridSize) * state.gridSize,
                  y: Math.round(position.y / state.gridSize) * state.gridSize
                }
              : position,
            size: ITEM_SIZES[type] || DEFAULT_ITEM_SIZE,
            content: null,
            required: false,
            points: undefined,
            explanation: undefined,
            zIndex: state.items.length + 1,
            selected: false
          };
          
          state.items.push(newItem);
          state.selectedItems = [newItem.id];
          
          // Update selection state
          state.items.forEach(item => {
            item.selected = item.id === newItem.id;
          });
        });
        
        get().saveState();
      },

      updateItem: (id: string, updates: Partial<CanvasItem>) => {
        set((state) => {
          const item = state.items.find(item => item.id === id);
          if (item) {
            Object.assign(item, updates);
          }
        });
        
        get().saveState();
      },

      deleteItem: (id: string) => {
        set((state) => {
          state.items = state.items.filter(item => item.id !== id);
          state.selectedItems = state.selectedItems.filter(itemId => itemId !== id);
        });
        
        get().saveState();
      },

      deleteSelectedItems: () => {
        const { selectedItems } = get();
        set((state) => {
          state.items = state.items.filter(item => !selectedItems.includes(item.id));
          state.selectedItems = [];
        });
        
        get().saveState();
      },

      selectItem: (id: string, multiSelect = false) => {
        set((state) => {
          if (multiSelect) {
            if (state.selectedItems.includes(id)) {
              state.selectedItems = state.selectedItems.filter(itemId => itemId !== id);
            } else {
              state.selectedItems.push(id);
            }
          } else {
            state.selectedItems = [id];
          }
          
          // Update selection state on items
          state.items.forEach(item => {
            item.selected = state.selectedItems.includes(item.id);
          });
        });
      },

      deselectAll: () => {
        set((state) => {
          state.selectedItems = [];
          state.items.forEach(item => {
            item.selected = false;
          });
        });
      },

      moveItem: (id: string, position: Position) => {
        set((state) => {
          const item = state.items.find(item => item.id === id);
          if (item) {
            item.position = state.snapToGrid 
              ? { 
                  x: Math.round(position.x / state.gridSize) * state.gridSize,
                  y: Math.round(position.y / state.gridSize) * state.gridSize
                }
              : position;
          }
        });
      },

      resizeItem: (id: string, size: Size) => {
        set((state) => {
          const item = state.items.find(item => item.id === id);
          if (item) {
            item.size = size;
          }
        });
        
        get().saveState();
      },

      duplicateItem: (id: string) => {
        set((state) => {
          const original = state.items.find(item => item.id === id);
          if (original) {
            const duplicate: CanvasItem = {
              ...original,
              id: generateId('item'),
              position: {
                x: original.position.x + 20,
                y: original.position.y + 20
              },
              zIndex: state.items.length + 1,
              selected: false
            };
            
            state.items.push(duplicate);
            state.selectedItems = [duplicate.id];
            
            // Update selection state
            state.items.forEach(item => {
              item.selected = item.id === duplicate.id;
            });
          }
        });
        
        get().saveState();
      },

      copySelectedItems: () => {
        const { items, selectedItems } = get();
        const selectedItemObjects = items.filter(item => selectedItems.includes(item.id));
        
        set((state) => {
          state.clipboardItems = selectedItemObjects.map(item => ({
            ...item,
            id: generateId('item'), // Generate new IDs for paste
            selected: false
          }));
        });
      },

      pasteItems: (position?: Position) => {
        const { clipboardItems } = get();
        if (clipboardItems.length === 0) return;
        
        set((state) => {
          const basePosition = position || { x: 50, y: 50 };
          
          const newItems = clipboardItems.map((item, index) => ({
            ...item,
            id: generateId('item'),
            position: {
              x: basePosition.x + (index * 20),
              y: basePosition.y + (index * 20)
            },
            zIndex: state.items.length + index + 1,
            selected: false
          }));
          
          state.items.push(...newItems);
          state.selectedItems = newItems.map(item => item.id);
          
          // Update selection state
          state.items.forEach(item => {
            item.selected = state.selectedItems.includes(item.id);
          });
        });
        
        get().saveState();
      },

      cutSelectedItems: () => {
        get().copySelectedItems();
        get().deleteSelectedItems();
      },

      saveState: () => {
        set((state) => {
          const currentState = {
            items: state.items,
            selectedItems: state.selectedItems,
            clipboardItems: state.clipboardItems,
            history: state.history,
            historyIndex: state.historyIndex,
            gridSize: state.gridSize,
            snapToGrid: state.snapToGrid,
            showGrid: state.showGrid,
            zoom: state.zoom,
            canvasSize: state.canvasSize
          };
          
          // Remove future history if we're not at the end
          if (state.historyIndex < state.history.length - 1) {
            state.history = state.history.slice(0, state.historyIndex + 1);
          }
          
          state.history.push(JSON.parse(JSON.stringify(currentState)));
          state.historyIndex = state.history.length - 1;
          
          // Limit history size
          if (state.history.length > 50) {
            state.history = state.history.slice(-50);
            state.historyIndex = 49;
          }
        });
      },

      undo: () => {
        const { history, historyIndex } = get();
        if (historyIndex > 0) {
          const prevState = history[historyIndex - 1];
          set((state) => {
            Object.assign(state, prevState);
            state.historyIndex = historyIndex - 1;
          });
        }
      },

      redo: () => {
        const { history, historyIndex } = get();
        if (historyIndex < history.length - 1) {
          const nextState = history[historyIndex + 1];
          set((state) => {
            Object.assign(state, nextState);
            state.historyIndex = historyIndex + 1;
          });
        }
      },

      canUndo: () => get().historyIndex > 0,
      canRedo: () => get().historyIndex < get().history.length - 1,

      setGridConfig: (config: Partial<GridConfig>) => {
        set((state) => {
          Object.assign(state.gridConfig, config);
          if (config.size) state.gridSize = config.size;
          if (config.snap !== undefined) state.snapToGrid = config.snap;
          if (config.visible !== undefined) state.showGrid = config.visible;
        });
      },

      setZoom: (zoom: number) => {
        set((state) => {
          state.zoom = Math.max(0.1, Math.min(3, zoom));
        });
      },

      setCanvasSize: (size: Size) => {
        set((state) => {
          state.canvasSize = size;
        });
      },

      resetCanvas: () => {
        set(() => ({
          ...createInitialState(),
          editorMode: EditorMode.EDIT,
          previewConfig: DEFAULT_PREVIEW_CONFIG,
          gridConfig: DEFAULT_GRID_CONFIG
        }));
      },

      setEditorMode: (mode: EditorMode) => {
        set((state) => {
          state.editorMode = mode;
        });
      },

      setPreviewConfig: (config: Partial<PreviewConfig>) => {
        set((state) => {
          Object.assign(state.previewConfig, config);
        });
      },

      loadCanvas: (canvasState: CanvasState) => {
        set(() => ({
          ...canvasState,
          editorMode: EditorMode.EDIT,
          previewConfig: DEFAULT_PREVIEW_CONFIG,
          gridConfig: DEFAULT_GRID_CONFIG
        }));
      },

      exportCanvas: () => {
        const state = get();
        return {
          items: state.items,
          selectedItems: state.selectedItems,
          clipboardItems: state.clipboardItems,
          history: state.history,
          historyIndex: state.historyIndex,
          gridSize: state.gridSize,
          snapToGrid: state.snapToGrid,
          showGrid: state.showGrid,
          zoom: state.zoom,
          canvasSize: state.canvasSize
        };
      }
    })),
    { name: 'canvas-store' }
  )
);