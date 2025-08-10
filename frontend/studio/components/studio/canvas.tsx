import { FC, useRef } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { useCanvasStore } from '@/store/canvas-store';
import { CanvasItem } from './canvas-item';
import { CanvasGrid } from './canvas-grid';
import { cn } from '@/lib/utils';
import { ItemType } from '@/types/quiz';

interface CanvasProps {
  className?: string;
}

export const Canvas: FC<CanvasProps> = ({ className }) => {
  const canvasRef = useRef<HTMLDivElement>(null);
  const {
    items,
    selectedItems,
    showGrid,
    zoom,
    canvasSize,
    snapToGrid,
    gridSize,
    addItem,
    deselectAll
  } = useCanvasStore();

  const { setNodeRef, isOver } = useDroppable({
    id: 'canvas',
    data: {
      accepts: ['palette-item']
    }
  });

  const handleCanvasClick = (e: React.MouseEvent) => {
    // Only deselect if clicking directly on canvas, not on items
    if (e.target === e.currentTarget) {
      deselectAll();
    }
  };

  const handleDrop = (event: any) => {
    const { active, over } = event;
    
    if (!over || over.id !== 'canvas') return;
    
    const data = active.data.current;
    if (!data || !data.isNew) return;
    
    const canvasRect = canvasRef.current?.getBoundingClientRect();
    if (!canvasRect) return;
    
    const position = {
      x: (event.activatorEvent.clientX - canvasRect.left) / zoom,
      y: (event.activatorEvent.clientY - canvasRect.top) / zoom
    };
    
    addItem(data.type as ItemType, position);
  };

  return (
    <div 
      className={cn(
        'flex-1 overflow-auto bg-gray-100 relative',
        isOver && 'bg-blue-50',
        className
      )}
      onClick={handleCanvasClick}
    >
      <div
        ref={(node) => {
          setNodeRef(node);
          if (canvasRef.current !== node) {
            canvasRef.current = node;
          }
        }}
        className="relative min-w-full min-h-full"
        style={{
          width: canvasSize.width * zoom,
          height: canvasSize.height * zoom,
          transform: `scale(${zoom})`,
          transformOrigin: '0 0'
        }}
      >
        {/* Grid Background */}
        {showGrid && (
          <CanvasGrid 
            size={gridSize}
            zoom={zoom}
            canvasSize={canvasSize}
          />
        )}
        
        {/* Drop Zone Indicator */}
        {isOver && (
          <div className="absolute inset-0 border-2 border-dashed border-blue-400 bg-blue-50/20 rounded-lg pointer-events-none z-10">
            <div className="flex items-center justify-center h-full">
              <div className="bg-blue-100 text-blue-700 px-4 py-2 rounded-lg font-medium">
                Drop block here to add to canvas
              </div>
            </div>
          </div>
        )}
        
        {/* Canvas Items */}
        <div className="absolute inset-0">
          {items.map((item) => (
            <CanvasItem
              key={item.id}
              item={item}
              isSelected={selectedItems.includes(item.id)}
              snapToGrid={snapToGrid}
              gridSize={gridSize}
            />
          ))}
        </div>
        
        {/* Selection Info */}
        {selectedItems.length > 0 && (
          <div className="absolute bottom-4 left-4 bg-white border border-gray-200 rounded-lg p-2 shadow-sm text-sm text-gray-600 pointer-events-none z-20">
            {selectedItems.length} item{selectedItems.length !== 1 ? 's' : ''} selected
          </div>
        )}
      </div>
    </div>
  );
};