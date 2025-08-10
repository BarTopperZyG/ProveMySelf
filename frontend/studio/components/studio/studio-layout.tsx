import { FC, useState } from 'react';
import { DndContext, DragEndEvent, DragOverEvent, closestCenter } from '@dnd-kit/core';
import { useCanvasStore } from '@/store/canvas-store';
import { BlockPalette } from './block-palette';
import { Canvas } from './canvas';
import { Toolbar } from './toolbar';
import { PropertiesPanel } from './properties-panel';
import { PreviewPanel } from './preview-panel';
import { EditorMode, ItemType } from '@/types/quiz';
import { cn } from '@/lib/utils';

interface StudioLayoutProps {
  className?: string;
}

export const StudioLayout: FC<StudioLayoutProps> = ({ className }) => {
  const [rightPanelWidth, setRightPanelWidth] = useState(300);
  const [isResizing, setIsResizing] = useState(false);
  
  const { 
    editorMode, 
    addItem, 
    moveItem,
    setEditorMode 
  } = useCanvasStore();

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over) return;

    const activeData = active.data.current;
    const overData = over.data.current;

    // Handle dropping palette items onto canvas
    if (activeData?.isNew && over.id === 'canvas') {
      const rect = (over.rect as any)?.current?.getBoundingClientRect();
      if (!rect) return;

      const position = {
        x: Math.max(0, (event.activatorEvent as MouseEvent).clientX - rect.left - 150),
        y: Math.max(0, (event.activatorEvent as MouseEvent).clientY - rect.top - 50)
      };

      addItem(activeData.type as ItemType, position);
    }

    // Handle moving existing items
    if (activeData?.type === 'canvas-item' && over.id === 'canvas') {
      const item = activeData.item;
      const rect = (over.rect as any)?.current?.getBoundingClientRect();
      
      if (rect && item) {
        const newPosition = {
          x: (event.activatorEvent as MouseEvent).clientX - rect.left - item.size.width / 2,
          y: (event.activatorEvent as MouseEvent).clientY - rect.top - item.size.height / 2
        };
        
        moveItem(item.id, newPosition);
      }
    }
  };

  const handleDragOver = (event: DragOverEvent) => {
    // Handle drag over events if needed
  };

  const handleResizeStart = (e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);

    const startX = e.clientX;
    const startWidth = rightPanelWidth;

    const handleMouseMove = (e: MouseEvent) => {
      const newWidth = Math.max(250, Math.min(500, startWidth - (e.clientX - startX)));
      setRightPanelWidth(newWidth);
    };

    const handleMouseUp = () => {
      setIsResizing(false);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  };

  return (
    <DndContext
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
      onDragOver={handleDragOver}
    >
      <div className={cn('flex h-screen bg-gray-50', className)}>
        {/* Left Panel - Block Palette */}
        <BlockPalette className="flex-shrink-0" />

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Toolbar */}
          <Toolbar 
            onModeChange={setEditorMode}
            currentMode={editorMode}
          />

          {/* Canvas/Preview Area */}
          <div className="flex-1 flex min-h-0">
            <div className="flex-1 relative">
              {editorMode === EditorMode.EDIT && <Canvas />}
              {editorMode === EditorMode.PREVIEW && <PreviewPanel />}
            </div>

            {/* Right Panel - Properties */}
            {editorMode === EditorMode.EDIT && (
              <>
                {/* Resize Handle */}
                <div
                  className={cn(
                    'w-1 bg-gray-300 cursor-col-resize hover:bg-blue-400 transition-colors',
                    isResizing && 'bg-blue-400'
                  )}
                  onMouseDown={handleResizeStart}
                />

                {/* Properties Panel */}
                <PropertiesPanel 
                  style={{ width: rightPanelWidth }}
                  className="flex-shrink-0 border-l border-gray-200"
                />
              </>
            )}
          </div>
        </div>
      </div>
    </DndContext>
  );
};