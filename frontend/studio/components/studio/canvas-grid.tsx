import { FC } from 'react';
import { Size } from '@/types/quiz';

interface CanvasGridProps {
  size: number;
  zoom: number;
  canvasSize: Size;
  color?: string;
}

export const CanvasGrid: FC<CanvasGridProps> = ({ 
  size, 
  zoom, 
  canvasSize, 
  color = '#e5e7eb' 
}) => {
  const gridSize = size * zoom;
  
  // Calculate number of lines needed
  const horizontalLines = Math.ceil(canvasSize.height / size);
  const verticalLines = Math.ceil(canvasSize.width / size);
  
  return (
    <svg
      className="absolute inset-0 pointer-events-none"
      width={canvasSize.width}
      height={canvasSize.height}
      style={{ zIndex: 0 }}
    >
      {/* Vertical lines */}
      {Array.from({ length: verticalLines + 1 }, (_, i) => (
        <line
          key={`v-${i}`}
          x1={i * size}
          y1={0}
          x2={i * size}
          y2={canvasSize.height}
          stroke={color}
          strokeWidth={1}
          opacity={0.5}
        />
      ))}
      
      {/* Horizontal lines */}
      {Array.from({ length: horizontalLines + 1 }, (_, i) => (
        <line
          key={`h-${i}`}
          x1={0}
          y1={i * size}
          x2={canvasSize.width}
          y2={i * size}
          stroke={color}
          strokeWidth={1}
          opacity={0.5}
        />
      ))}
      
      {/* Major grid lines every 5th line */}
      {Array.from({ length: Math.ceil(verticalLines / 5) + 1 }, (_, i) => (
        <line
          key={`v-major-${i}`}
          x1={i * size * 5}
          y1={0}
          x2={i * size * 5}
          y2={canvasSize.height}
          stroke={color}
          strokeWidth={2}
          opacity={0.8}
        />
      ))}
      
      {Array.from({ length: Math.ceil(horizontalLines / 5) + 1 }, (_, i) => (
        <line
          key={`h-major-${i}`}
          x1={0}
          y1={i * size * 5}
          x2={canvasSize.width}
          y2={i * size * 5}
          stroke={color}
          strokeWidth={2}
          opacity={0.8}
        />
      ))}
    </svg>
  );
};