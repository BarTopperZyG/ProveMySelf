import { CanvasItem, ItemType, AssetReference } from '@/types/quiz';

export interface AssetFile {
  id: string;
  name: string;
  type: 'image' | 'video' | 'audio' | 'document';
  mimeType: string;
  size: number;
  url: string;
  blob?: Blob;
  metadata?: {
    width?: number;
    height?: number;
    duration?: number;
    pages?: number;
  };
  createdAt: string;
  usageCount: number;
}

export interface AssetBundle {
  assets: AssetFile[];
  manifest: {
    version: string;
    created: string;
    totalSize: number;
    totalFiles: number;
  };
}

export class AssetManager {
  private assets: Map<string, AssetFile> = new Map();
  private urlToAssetMap: Map<string, string> = new Map();

  /**
   * Extract all asset references from canvas items
   */
  extractAssetReferences(items: CanvasItem[]): AssetReference[] {
    const references: AssetReference[] = [];

    for (const item of items) {
      switch (item.type) {
        case ItemType.MEDIA:
          if (item.content?.url) {
            references.push({
              id: `asset_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              type: this.getAssetTypeFromUrl(item.content.url),
              url: item.content.url,
              itemId: item.id,
              property: 'content.url',
              required: true
            });
          }
          break;

        case ItemType.HOTSPOT:
          if (item.content?.imageUrl) {
            references.push({
              id: `asset_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              type: 'image',
              url: item.content.imageUrl,
              itemId: item.id,
              property: 'content.imageUrl',
              required: true
            });
          }
          break;

        case ItemType.CHOICE:
        case ItemType.MULTI_CHOICE:
          if (item.content?.choices) {
            for (const [choiceIndex, choice] of item.content.choices.entries()) {
              if (choice.imageUrl) {
                references.push({
                  id: `asset_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
                  type: 'image',
                  url: choice.imageUrl,
                  itemId: item.id,
                  property: `content.choices[${choiceIndex}].imageUrl`,
                  required: false
                });
              }
            }
          }
          break;

        default:
          // Check for any embedded media in text content
          const textContent = item.title + ' ' + (item.explanation || '');
          const urlMatches = textContent.match(/https?:\/\/[^\s]+\.(jpg|jpeg|png|gif|webp|mp4|mp3|wav|pdf)/gi);
          if (urlMatches) {
            for (const url of urlMatches) {
              references.push({
                id: `asset_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
                type: this.getAssetTypeFromUrl(url),
                url,
                itemId: item.id,
                property: 'embedded',
                required: false
              });
            }
          }
          break;
      }
    }

    return references;
  }

  /**
   * Download and process all assets from URLs
   */
  async downloadAssets(references: AssetReference[]): Promise<AssetFile[]> {
    const downloadPromises = references.map(async (ref) => {
      try {
        // Skip if asset already exists
        const existingAssetId = this.urlToAssetMap.get(ref.url);
        if (existingAssetId && this.assets.has(existingAssetId)) {
          const asset = this.assets.get(existingAssetId)!;
          asset.usageCount++;
          return asset;
        }

        // Download the asset
        const response = await fetch(ref.url);
        if (!response.ok) {
          throw new Error(`Failed to fetch ${ref.url}: ${response.statusText}`);
        }

        const blob = await response.blob();
        const filename = this.extractFilenameFromUrl(ref.url);
        
        // Create asset file
        const asset: AssetFile = {
          id: ref.id,
          name: filename,
          type: ref.type,
          mimeType: blob.type || this.getMimeTypeFromExtension(filename),
          size: blob.size,
          url: ref.url,
          blob,
          metadata: await this.extractMetadata(blob, ref.type),
          createdAt: new Date().toISOString(),
          usageCount: 1
        };

        // Store in maps
        this.assets.set(asset.id, asset);
        this.urlToAssetMap.set(ref.url, asset.id);

        return asset;
      } catch (error) {
        console.warn(`Failed to download asset from ${ref.url}:`, error);
        // Return a placeholder asset for missing resources
        return {
          id: ref.id,
          name: this.extractFilenameFromUrl(ref.url),
          type: ref.type,
          mimeType: 'application/octet-stream',
          size: 0,
          url: ref.url,
          createdAt: new Date().toISOString(),
          usageCount: 1
        };
      }
    });

    return Promise.all(downloadPromises);
  }

  /**
   * Create a complete asset bundle with all files
   */
  async createAssetBundle(items: CanvasItem[]): Promise<AssetBundle> {
    const references = this.extractAssetReferences(items);
    const assets = await this.downloadAssets(references);

    const totalSize = assets.reduce((sum, asset) => sum + asset.size, 0);

    return {
      assets,
      manifest: {
        version: '1.0.0',
        created: new Date().toISOString(),
        totalSize,
        totalFiles: assets.length
      }
    };
  }

  /**
   * Generate a ZIP-like bundle as a Blob
   */
  async createDownloadableBundle(
    items: CanvasItem[],
    uiJson: any[],
    quizJson: any[],
    metadata: any
  ): Promise<Blob> {
    const assetBundle = await this.createAssetBundle(items);
    
    // Create the bundle structure
    const bundle = {
      metadata,
      ui: uiJson,
      quiz: quizJson,
      assets: assetBundle.assets.map(asset => ({
        id: asset.id,
        name: asset.name,
        type: asset.type,
        mimeType: asset.mimeType,
        size: asset.size,
        url: asset.url,
        metadata: asset.metadata
      })),
      manifest: assetBundle.manifest
    };

    // For now, return as JSON blob. In a full implementation,
    // this would create a ZIP file with separate asset files.
    const bundleJson = JSON.stringify(bundle, null, 2);
    return new Blob([bundleJson], { type: 'application/json' });
  }

  /**
   * Replace asset URLs in items with local references
   */
  localizeAssetReferences(items: CanvasItem[], assets: AssetFile[]): CanvasItem[] {
    const urlToLocalMap = new Map<string, string>();
    
    // Build URL to local path mapping
    for (const asset of assets) {
      urlToLocalMap.set(asset.url, `./assets/${asset.name}`);
    }

    return items.map(item => {
      const updatedItem = { ...item };

      switch (item.type) {
        case ItemType.MEDIA:
          if (item.content?.url && urlToLocalMap.has(item.content.url)) {
            updatedItem.content = {
              ...item.content,
              url: urlToLocalMap.get(item.content.url)!
            };
          }
          break;

        case ItemType.HOTSPOT:
          if (item.content?.imageUrl && urlToLocalMap.has(item.content.imageUrl)) {
            updatedItem.content = {
              ...item.content,
              imageUrl: urlToLocalMap.get(item.content.imageUrl)!
            };
          }
          break;

        case ItemType.CHOICE:
        case ItemType.MULTI_CHOICE:
          if (item.content?.choices) {
            updatedItem.content = {
              ...item.content,
              choices: item.content.choices.map((choice: any) => {
                if (choice.imageUrl && urlToLocalMap.has(choice.imageUrl)) {
                  return {
                    ...choice,
                    imageUrl: urlToLocalMap.get(choice.imageUrl)!
                  };
                }
                return choice;
              })
            };
          }
          break;
      }

      return updatedItem;
    });
  }

  /**
   * Get asset statistics
   */
  getAssetStats(): {
    totalAssets: number;
    totalSize: number;
    byType: Record<string, { count: number; size: number }>;
  } {
    const stats = {
      totalAssets: this.assets.size,
      totalSize: 0,
      byType: {} as Record<string, { count: number; size: number }>
    };

    for (const asset of this.assets.values()) {
      stats.totalSize += asset.size;
      
      if (!stats.byType[asset.type]) {
        stats.byType[asset.type] = { count: 0, size: 0 };
      }
      stats.byType[asset.type].count++;
      stats.byType[asset.type].size += asset.size;
    }

    return stats;
  }

  /**
   * Helper methods
   */
  private getAssetTypeFromUrl(url: string): 'image' | 'video' | 'audio' | 'document' {
    const extension = url.split('.').pop()?.toLowerCase() || '';
    
    if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp'].includes(extension)) {
      return 'image';
    } else if (['mp4', 'webm', 'avi', 'mov', 'wmv'].includes(extension)) {
      return 'video';
    } else if (['mp3', 'wav', 'ogg', 'aac', 'm4a'].includes(extension)) {
      return 'audio';
    } else {
      return 'document';
    }
  }

  private getMimeTypeFromExtension(filename: string): string {
    const extension = filename.split('.').pop()?.toLowerCase() || '';
    
    const mimeTypes: Record<string, string> = {
      'jpg': 'image/jpeg',
      'jpeg': 'image/jpeg',
      'png': 'image/png',
      'gif': 'image/gif',
      'webp': 'image/webp',
      'svg': 'image/svg+xml',
      'mp4': 'video/mp4',
      'webm': 'video/webm',
      'mp3': 'audio/mpeg',
      'wav': 'audio/wav',
      'ogg': 'audio/ogg',
      'pdf': 'application/pdf'
    };

    return mimeTypes[extension] || 'application/octet-stream';
  }

  private extractFilenameFromUrl(url: string): string {
    try {
      const urlObj = new URL(url);
      const pathname = urlObj.pathname;
      return pathname.substring(pathname.lastIndexOf('/') + 1) || 'unnamed_asset';
    } catch {
      return 'unnamed_asset';
    }
  }

  private async extractMetadata(blob: Blob, type: AssetFile['type']): Promise<AssetFile['metadata']> {
    const metadata: AssetFile['metadata'] = {};

    try {
      if (type === 'image') {
        // Extract image dimensions
        const imageUrl = URL.createObjectURL(blob);
        const img = new Image();
        
        await new Promise((resolve, reject) => {
          img.onload = resolve;
          img.onerror = reject;
          img.src = imageUrl;
        });

        metadata.width = img.width;
        metadata.height = img.height;
        URL.revokeObjectURL(imageUrl);
      } else if (type === 'video') {
        // Extract video metadata
        const videoUrl = URL.createObjectURL(blob);
        const video = document.createElement('video');
        
        await new Promise((resolve, reject) => {
          video.onloadedmetadata = resolve;
          video.onerror = reject;
          video.src = videoUrl;
        });

        metadata.width = video.videoWidth;
        metadata.height = video.videoHeight;
        metadata.duration = video.duration;
        URL.revokeObjectURL(videoUrl);
      } else if (type === 'audio') {
        // Extract audio duration
        const audioUrl = URL.createObjectURL(blob);
        const audio = new Audio();
        
        await new Promise((resolve, reject) => {
          audio.onloadedmetadata = resolve;
          audio.onerror = reject;
          audio.src = audioUrl;
        });

        metadata.duration = audio.duration;
        URL.revokeObjectURL(audioUrl);
      }
    } catch (error) {
      console.warn('Failed to extract metadata:', error);
    }

    return metadata;
  }

  /**
   * Clear all assets
   */
  clear(): void {
    this.assets.clear();
    this.urlToAssetMap.clear();
  }

  /**
   * Get all assets
   */
  getAllAssets(): AssetFile[] {
    return Array.from(this.assets.values());
  }

  /**
   * Get asset by ID
   */
  getAsset(id: string): AssetFile | undefined {
    return this.assets.get(id);
  }
}