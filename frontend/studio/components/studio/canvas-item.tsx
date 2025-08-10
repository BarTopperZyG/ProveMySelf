import { FC, useState, useRef } from 'react';
import { useDraggable } from '@dnd-kit/core';
import { CanvasItem as CanvasItemType, ItemType } from '@/types/quiz';
import { useCanvasStore } from '@/store/canvas-store';
import { cn } from '@/lib/utils';
import { 
  Type, 
  Image, 
  CheckCircle, 
  CheckSquare, 
  FileText, 
  List, 
  MousePointer,
  Move,
  RotateCcw,
  Trash2
} from 'lucide-react';

interface CanvasItemProps {
  item: CanvasItemType;
  isSelected: boolean;
  snapToGrid: boolean;
  gridSize: number;
}

const ITEM_ICONS: Record<ItemType, React.ReactNode> = {
  [ItemType.TITLE]: <Type className="w-4 h-4" />,
  [ItemType.MEDIA]: <Image className="w-4 h-4" />,
  [ItemType.CHOICE]: <CheckCircle className="w-4 h-4" />,
  [ItemType.MULTI_CHOICE]: <CheckSquare className="w-4 h-4" />,
  [ItemType.TEXT_ENTRY]: <FileText className="w-4 h-4" />,
  [ItemType.ORDERING]: <List className="w-4 h-4" />,
  [ItemType.HOTSPOT]: <MousePointer className="w-4 h-4" />
};

const ITEM_COLORS: Record<ItemType, string> = {
  [ItemType.TITLE]: 'border-blue-300 bg-blue-50',
  [ItemType.MEDIA]: 'border-green-300 bg-green-50',
  [ItemType.CHOICE]: 'border-purple-300 bg-purple-50',
  [ItemType.MULTI_CHOICE]: 'border-indigo-300 bg-indigo-50',
  [ItemType.TEXT_ENTRY]: 'border-yellow-300 bg-yellow-50',
  [ItemType.ORDERING]: 'border-orange-300 bg-orange-50',
  [ItemType.HOTSPOT]: 'border-red-300 bg-red-50'
};

interface ResizeHandleProps {
  position: 'nw' | 'ne' | 'sw' | 'se' | 'n' | 'e' | 's' | 'w';
  onResize: (direction: string, deltaX: number, deltaY: number) => void;
}

const ResizeHandle: FC<ResizeHandleProps> = ({ position, onResize }) => {
  const [isDragging, setIsDragging] = useState(false);
  const startPos = useRef({ x: 0, y: 0 });

  const handleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
    startPos.current = { x: e.clientX, y: e.clientY };

    const handleMouseMove = (e: MouseEvent) => {
      const deltaX = e.clientX - startPos.current.x;
      const deltaY = e.clientY - startPos.current.y;
      onResize(position, deltaX, deltaY);
      startPos.current = { x: e.clientX, y: e.clientY };
    };

    const handleMouseUp = () => {
      setIsDragging(false);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  };

  const getCursorStyle = () => {
    switch (position) {
      case 'nw':
      case 'se':
        return 'cursor-nw-resize';
      case 'ne':
      case 'sw':
        return 'cursor-ne-resize';
      case 'n':
      case 's':
        return 'cursor-ns-resize';
      case 'e':
      case 'w':
        return 'cursor-ew-resize';
      default:
        return 'cursor-default';
    }
  };

  const getPositionStyle = () => {
    const base = 'absolute w-2 h-2 bg-blue-500 border border-white rounded-sm';
    switch (position) {
      case 'nw':
        return `${base} -top-1 -left-1`;
      case 'ne':
        return `${base} -top-1 -right-1`;
      case 'sw':
        return `${base} -bottom-1 -left-1`;
      case 'se':
        return `${base} -bottom-1 -right-1`;
      case 'n':
        return `${base} -top-1 left-1/2 -translate-x-1/2`;
      case 'e':
        return `${base} top-1/2 -right-1 -translate-y-1/2`;
      case 's':
        return `${base} -bottom-1 left-1/2 -translate-x-1/2`;
      case 'w':
        return `${base} top-1/2 -left-1 -translate-y-1/2`;
      default:
        return base;
    }
  };

  return (
    <div
      className={cn(getPositionStyle(), getCursorStyle())}
      onMouseDown={handleMouseDown}
      style={{ zIndex: 1000 }}
    />
  );
};

export const CanvasItem: FC<CanvasItemProps> = ({ 
  item, 
  isSelected, 
  snapToGrid, 
  gridSize 
}) => {
  const { 
    selectItem, 
    moveItem, 
    resizeItem, 
    deleteItem, 
    duplicateItem 
  } = useCanvasStore();

  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: item.id,
    data: {
      type: 'canvas-item',
      item
    }
  });

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    selectItem(item.id, e.ctrlKey || e.metaKey);
  };

  const handleResize = (direction: string, deltaX: number, deltaY: number) => {
    let newWidth = item.size.width;
    let newHeight = item.size.height;
    let newX = item.position.x;
    let newY = item.position.y;

    switch (direction) {
      case 'nw':
        newWidth = Math.max(50, item.size.width - deltaX);
        newHeight = Math.max(30, item.size.height - deltaY);
        newX = item.position.x + deltaX;
        newY = item.position.y + deltaY;
        break;
      case 'ne':
        newWidth = Math.max(50, item.size.width + deltaX);
        newHeight = Math.max(30, item.size.height - deltaY);
        newY = item.position.y + deltaY;
        break;
      case 'sw':
        newWidth = Math.max(50, item.size.width - deltaX);
        newHeight = Math.max(30, item.size.height + deltaY);
        newX = item.position.x + deltaX;
        break;
      case 'se':
        newWidth = Math.max(50, item.size.width + deltaX);
        newHeight = Math.max(30, item.size.height + deltaY);
        break;
      case 'n':
        newHeight = Math.max(30, item.size.height - deltaY);
        newY = item.position.y + deltaY;
        break;
      case 'e':
        newWidth = Math.max(50, item.size.width + deltaX);
        break;
      case 's':
        newHeight = Math.max(30, item.size.height + deltaY);
        break;
      case 'w':
        newWidth = Math.max(50, item.size.width - deltaX);
        newX = item.position.x + deltaX;
        break;
    }

    if (snapToGrid) {
      newX = Math.round(newX / gridSize) * gridSize;
      newY = Math.round(newY / gridSize) * gridSize;
      newWidth = Math.round(newWidth / gridSize) * gridSize;
      newHeight = Math.round(newHeight / gridSize) * gridSize;
    }

    moveItem(item.id, { x: newX, y: newY });
    resizeItem(item.id, { width: newWidth, height: newHeight });
  };

  const style = transform ? {
    transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
  } : undefined;

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'absolute border-2 rounded-lg cursor-pointer select-none',
        ITEM_COLORS[item.type],
        isSelected 
          ? 'border-blue-500 shadow-lg ring-2 ring-blue-200' 
          : 'border-gray-300 hover:border-gray-400',
        isDragging && 'opacity-50 z-50'
      )}
      style={{
        left: item.position.x,
        top: item.position.y,
        width: item.size.width,
        height: item.size.height,
        zIndex: item.zIndex,
        ...style
      }}
      onClick={handleClick}
      {...attributes}
      {...listeners}
    >
      {/* Item Header */}
      <div className="flex items-center justify-between p-2 border-b border-gray-200 bg-white rounded-t-md">
        <div className="flex items-center gap-2">
          {ITEM_ICONS[item.type]}
          <span className="text-sm font-medium text-gray-700">
            {item.title}
          </span>
        </div>
        
        {isSelected && (
          <div className="flex items-center gap-1">
            <button
              className="p-1 text-gray-500 hover:text-blue-600 hover:bg-blue-100 rounded"
              onClick={(e) => {
                e.stopPropagation();
                duplicateItem(item.id);
              }}
              title="Duplicate"
            >
              <RotateCcw className="w-3 h-3" />
            </button>
            <button
              className="p-1 text-gray-500 hover:text-red-600 hover:bg-red-100 rounded"
              onClick={(e) => {
                e.stopPropagation();
                deleteItem(item.id);
              }}
              title="Delete"
            >
              <Trash2 className="w-3 h-3" />
            </button>
          </div>
        )}
      </div>

      {/* Item Content */}
      <div className="p-3 flex-1 overflow-hidden">
        <div className="text-sm text-gray-600">
          {item.type === ItemType.TITLE && 'Click to edit title content'}
          {item.type === ItemType.MEDIA && 'Click to add media content'}
          {item.type === ItemType.CHOICE && 'Click to configure choice options'}
          {item.type === ItemType.MULTI_CHOICE && 'Click to configure multiple choice options'}
          {item.type === ItemType.TEXT_ENTRY && 'Click to configure text input'}
          {item.type === ItemType.ORDERING && 'Click to configure ordering items'}
          {item.type === ItemType.HOTSPOT && 'Click to configure hotspot areas'}
        </div>
        
        {item.points && (
          <div className="mt-2 text-xs text-gray-500">
            Points: {item.points}
          </div>
        )}
        
        {item.required && (
          <div className="mt-1 text-xs text-red-600 font-medium">
            Required
          </div>
        )}
      </div>

      {/* Resize Handles */}
      {isSelected && (
        <>
          <ResizeHandle position="nw" onResize={handleResize} />
          <ResizeHandle position="ne" onResize={handleResize} />
          <ResizeHandle position="sw" onResize={handleResize} />
          <ResizeHandle position="se" onResize={handleResize} />
          <ResizeHandle position="n" onResize={handleResize} />
          <ResizeHandle position="e" onResize={handleResize} />
          <ResizeHandle position="s" onResize={handleResize} />
          <ResizeHandle position="w" onResize={handleResize} />
        </>
      )}

      {/* Move Handle */}
      {isSelected && (
        <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-blue-500 text-white p-1 rounded text-xs flex items-center gap-1">
          <Move className="w-3 h-3" />
          <span>Drag</span>
        </div>
      )}
    </div>
  );
};