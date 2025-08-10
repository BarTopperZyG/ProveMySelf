import { FC, useState } from 'react';
import { useCanvasStore } from '@/store/canvas-store';
import { Button } from '@/components/ui/button';
import { EditorMode } from '@/types/quiz';
import { cn } from '@/lib/utils';
import {
  Edit,
  Eye,
  Save,
  Download,
  Upload,
  Undo2,
  Redo2,
  Copy,
  Scissors,
  Clipboard,
  Trash2,
  Grid3x3,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  Settings,
  Play
} from 'lucide-react';
import { useHotkeys } from 'react-hotkeys-hook';
import { ExportDialog } from './export-dialog';

interface ToolbarProps {
  onModeChange: (mode: EditorMode) => void;
  currentMode: EditorMode;
  className?: string;
}

export const Toolbar: FC<ToolbarProps> = ({ 
  onModeChange, 
  currentMode, 
  className 
}) => {
  const [showExportDialog, setShowExportDialog] = useState(false);
  
  const {
    selectedItems,
    canUndo,
    canRedo,
    undo,
    redo,
    copySelectedItems,
    cutSelectedItems,
    pasteItems,
    deleteSelectedItems,
    duplicateItem,
    zoom,
    setZoom,
    showGrid,
    snapToGrid,
    setGridConfig,
    resetCanvas
  } = useCanvasStore();

  // Keyboard shortcuts
  useHotkeys('ctrl+z,cmd+z', () => undo(), { enabled: canUndo() });
  useHotkeys('ctrl+y,cmd+y,ctrl+shift+z,cmd+shift+z', () => redo(), { enabled: canRedo() });
  useHotkeys('ctrl+c,cmd+c', () => copySelectedItems(), { enabled: selectedItems.length > 0 });
  useHotkeys('ctrl+x,cmd+x', () => cutSelectedItems(), { enabled: selectedItems.length > 0 });
  useHotkeys('ctrl+v,cmd+v', () => pasteItems());
  useHotkeys('delete,backspace', () => deleteSelectedItems(), { enabled: selectedItems.length > 0 });
  useHotkeys('ctrl+d,cmd+d', () => {
    if (selectedItems.length === 1) {
      duplicateItem(selectedItems[0]);
    }
  }, { enabled: selectedItems.length === 1, preventDefault: true });

  const handleZoomIn = () => setZoom(Math.min(3, zoom + 0.1));
  const handleZoomOut = () => setZoom(Math.max(0.1, zoom - 0.1));
  const handleResetZoom = () => setZoom(1);

  const toggleGrid = () => setGridConfig({ visible: !showGrid });
  const toggleSnap = () => setGridConfig({ snap: !snapToGrid });

  const isEditMode = currentMode === EditorMode.EDIT;
  const isPreviewMode = currentMode === EditorMode.PREVIEW;
  const hasSelection = selectedItems.length > 0;

  return (
    <div className={cn(
      'flex items-center justify-between px-4 py-2 bg-white border-b border-gray-200 shadow-sm',
      className
    )}>
      {/* Left Section - Mode Toggle */}
      <div className="flex items-center gap-2">
        <div className="flex items-center bg-gray-100 rounded-lg p-1">
          <Button
            variant={isEditMode ? "default" : "ghost"}
            size="sm"
            onClick={() => onModeChange(EditorMode.EDIT)}
            className="gap-2"
          >
            <Edit className="w-4 h-4" />
            Edit
          </Button>
          <Button
            variant={isPreviewMode ? "default" : "ghost"}
            size="sm"
            onClick={() => onModeChange(EditorMode.PREVIEW)}
            className="gap-2"
          >
            <Eye className="w-4 h-4" />
            Preview
          </Button>
        </div>

        {/* Separator */}
        <div className="w-px h-6 bg-gray-300 mx-2" />

        {/* File Operations */}
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" className="gap-2">
            <Save className="w-4 h-4" />
            Save
          </Button>
          <Button variant="ghost" size="sm" className="gap-2">
            <Upload className="w-4 h-4" />
            Import
          </Button>
          <Button 
            variant="ghost" 
            size="sm" 
            className="gap-2"
            onClick={() => setShowExportDialog(true)}
          >
            <Download className="w-4 h-4" />
            Export
          </Button>
        </div>
      </div>

      {/* Center Section - Edit Tools (only in edit mode) */}
      {isEditMode && (
        <div className="flex items-center gap-2">
          {/* History */}
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={undo}
              disabled={!canUndo()}
              title="Undo (Ctrl+Z)"
            >
              <Undo2 className="w-4 h-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={redo}
              disabled={!canRedo()}
              title="Redo (Ctrl+Y)"
            >
              <Redo2 className="w-4 h-4" />
            </Button>
          </div>

          {/* Separator */}
          <div className="w-px h-6 bg-gray-300" />

          {/* Clipboard Operations */}
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={copySelectedItems}
              disabled={!hasSelection}
              title="Copy (Ctrl+C)"
            >
              <Copy className="w-4 h-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={cutSelectedItems}
              disabled={!hasSelection}
              title="Cut (Ctrl+X)"
            >
              <Scissors className="w-4 h-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => pasteItems()}
              title="Paste (Ctrl+V)"
            >
              <Clipboard className="w-4 h-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={deleteSelectedItems}
              disabled={!hasSelection}
              title="Delete (Del)"
            >
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>

          {/* Separator */}
          <div className="w-px h-6 bg-gray-300" />

          {/* Grid Controls */}
          <div className="flex items-center gap-1">
            <Button
              variant={showGrid ? "default" : "ghost"}
              size="sm"
              onClick={toggleGrid}
              title="Toggle Grid"
            >
              <Grid3x3 className="w-4 h-4" />
            </Button>
            <Button
              variant={snapToGrid ? "default" : "ghost"}
              size="sm"
              onClick={toggleSnap}
              title="Snap to Grid"
              className="gap-1"
            >
              <Grid3x3 className="w-3 h-3" />
              Snap
            </Button>
          </div>

          {/* Separator */}
          <div className="w-px h-6 bg-gray-300" />

          {/* Zoom Controls */}
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleZoomOut}
              disabled={zoom <= 0.1}
              title="Zoom Out"
            >
              <ZoomOut className="w-4 h-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleResetZoom}
              title="Reset Zoom (100%)"
              className="min-w-16 text-xs"
            >
              {Math.round(zoom * 100)}%
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleZoomIn}
              disabled={zoom >= 3}
              title="Zoom In"
            >
              <ZoomIn className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Preview Controls (only in preview mode) */}
      {isPreviewMode && (
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="gap-2">
            <Play className="w-4 h-4" />
            Start Quiz
          </Button>
          <Button variant="ghost" size="sm" className="gap-2">
            <RotateCcw className="w-4 h-4" />
            Reset
          </Button>
        </div>
      )}

      {/* Right Section - Settings */}
      <div className="flex items-center gap-2">
        {selectedItems.length > 0 && (
          <div className="text-sm text-gray-600 mr-4">
            {selectedItems.length} item{selectedItems.length !== 1 ? 's' : ''} selected
          </div>
        )}
        
        <Button variant="ghost" size="sm">
          <Settings className="w-4 h-4" />
        </Button>
      </div>

      {/* Export Dialog */}
      <ExportDialog
        isOpen={showExportDialog}
        onClose={() => setShowExportDialog(false)}
      />
    </div>
  );
};