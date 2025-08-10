import { z } from 'zod';

// Item types matching backend
export enum ItemType {
  TITLE = 'title',
  MEDIA = 'media',
  CHOICE = 'choice',
  MULTI_CHOICE = 'multi_choice',
  TEXT_ENTRY = 'text_entry',
  ORDERING = 'ordering',
  HOTSPOT = 'hotspot'
}

// Canvas position and size
export interface Position {
  x: number;
  y: number;
}

export interface Size {
  width: number;
  height: number;
}

export interface Bounds extends Position, Size {}

// Canvas item representation
export interface CanvasItem {
  id: string;
  type: ItemType;
  title: string;
  position: Position;
  size: Size;
  content?: any;
  required: boolean;
  points?: number;
  explanation?: string;
  zIndex: number;
  selected: boolean;
}

// Choice-related schemas
export const ChoiceSchema = z.object({
  id: z.string(),
  text: z.string().min(1).max(500),
  correct: z.boolean()
});

export const ChoiceContentSchema = z.object({
  choices: z.array(ChoiceSchema).min(1).max(10)
});

// Media-related schemas
export const MediaContentSchema = z.object({
  url: z.string().url(),
  mediaType: z.enum(['image', 'video', 'audio']),
  altText: z.string().max(200).optional(),
  caption: z.string().max(500).optional(),
  autoplay: z.boolean().default(false),
  showControls: z.boolean().default(true)
});

// Text entry schemas
export const TextEntryContentSchema = z.object({
  maxLength: z.number().min(1).max(10000).optional(),
  placeholder: z.string().max(100).optional(),
  multiline: z.boolean().default(false),
  correctAnswer: z.string().max(10000).optional()
});

// Ordering schemas
export const OrderingItemSchema = z.object({
  id: z.string(),
  text: z.string().min(1).max(500),
  correctOrder: z.number().min(1)
});

export const OrderingContentSchema = z.object({
  items: z.array(OrderingItemSchema).min(2).max(10)
});

// Hotspot schemas
export const HotspotSchema = z.object({
  id: z.string(),
  shape: z.enum(['rectangle', 'circle', 'polygon']),
  coords: z.array(z.number()).min(2),
  correct: z.boolean(),
  feedback: z.string().max(200).optional()
});

export const HotspotContentSchema = z.object({
  imageUrl: z.string().url(),
  altText: z.string().max(200).optional(),
  hotspots: z.array(HotspotSchema).min(1).max(20)
});

// Canvas state
export interface CanvasState {
  items: CanvasItem[];
  selectedItems: string[];
  clipboardItems: CanvasItem[];
  history: CanvasState[];
  historyIndex: number;
  gridSize: number;
  snapToGrid: boolean;
  showGrid: boolean;
  zoom: number;
  canvasSize: Size;
}

// Editor modes
export enum EditorMode {
  EDIT = 'edit',
  PREVIEW = 'preview',
  PROPERTIES = 'properties'
}

// Quiz project state
export interface QuizProject {
  id: string;
  title: string;
  description?: string;
  tags: string[];
  createdAt: string;
  updatedAt: string;
  publishedAt?: string;
}

// Export formats
export interface AdaptiveCard {
  type: string;
  version: string;
  body: any[];
}

export interface QTIItem {
  identifier: string;
  title: string;
  responseDeclaration?: any;
  outcomeDeclaration?: any;
  itemBody: any;
  responseProcessing?: any;
}

export interface QuizBundle {
  metadata: {
    title: string;
    description?: string;
    version: string;
    createdAt: string;
  };
  ui: AdaptiveCard;
  qti: {
    version: string;
    items: QTIItem[];
  };
  assets: {
    [key: string]: string; // filename -> base64 data
  };
}

// Property panel schemas
export const ItemPropertiesSchema = z.object({
  title: z.string().min(1).max(500),
  required: z.boolean(),
  points: z.number().min(0).max(1000).optional(),
  explanation: z.string().max(1000).optional()
});

// Validation types
export type Choice = z.infer<typeof ChoiceSchema>;
export type ChoiceContent = z.infer<typeof ChoiceContentSchema>;
export type MediaContent = z.infer<typeof MediaContentSchema>;
export type TextEntryContent = z.infer<typeof TextEntryContentSchema>;
export type OrderingItem = z.infer<typeof OrderingItemSchema>;
export type OrderingContent = z.infer<typeof OrderingContentSchema>;
export type Hotspot = z.infer<typeof HotspotSchema>;
export type HotspotContent = z.infer<typeof HotspotContentSchema>;
export type ItemProperties = z.infer<typeof ItemPropertiesSchema>;

// Content union type
export type ItemContent = 
  | ChoiceContent 
  | MediaContent 
  | TextEntryContent 
  | OrderingContent 
  | HotspotContent 
  | null;

// Drag and drop types
export interface DragItem {
  type: ItemType;
  isNew: boolean;
  item?: CanvasItem;
}

// Grid configuration
export interface GridConfig {
  size: number;
  snap: boolean;
  visible: boolean;
  color: string;
}

// Preview configuration
export interface PreviewConfig {
  mode: 'desktop' | 'tablet' | 'mobile';
  showNavigation: boolean;
  autoAdvance: boolean;
  allowBack: boolean;
}

// Export configuration
export interface ExportConfig {
  includeAssets: boolean;
  format: 'json' | 'zip';
  version: string;
  metadata: {
    title: string;
    description?: string;
    author?: string;
  };
}