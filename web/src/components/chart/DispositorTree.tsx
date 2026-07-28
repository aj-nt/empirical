import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import type { DispositorTree } from '../../lib/types';
import { planetGlyph } from '../../lib/astrology';

interface DispositorTreeProps {
  data: DispositorTree;
}

interface HierarchyDatum {
  name: string;
  sign: string;
  isFinal: boolean;
  inLoop: boolean;
  mutualReception: boolean;
  children?: HierarchyDatum[];
}

export function DispositorTree({ data }: DispositorTreeProps) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current || !data.nodes.length) return;

    // Clear previous render
    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    // Build lookup
    const nodeMap = new Map(data.nodes.map((n) => [n.planet, n]));

    // Determine mutual reception set for quick lookup
    const mrSet = new Set<string>();
    if (data.mutual_receptions) {
      for (const pair of data.mutual_receptions) {
        for (const p of pair) mrSet.add(p);
      }
    }

    // Build hierarchy: find roots
    // A planet is a root if it's a final dispositor (disposes itself) OR no other node has it as their dispositor
    const disposedBy = new Map<string, string>();
    for (const n of data.nodes) {
      disposedBy.set(n.dispositor, n.planet); // dispositor -> child (last one wins, but we just need existence)
    }

    const roots = data.nodes.filter((n) => {
      if (n.is_final) return true;
      // Also root if no other node lists this planet as their dispositor
      // (i.e., nothing points to this planet)
      const isReferenced = data.nodes.some((m) => m.dispositor === n.planet && m.planet !== n.planet);
      return !isReferenced;
    });

    // Deduplicate roots by planet name
    const seenRoots = new Set<string>();
    const uniqueRoots = roots.filter((n) => {
      if (seenRoots.has(n.planet)) return false;
      seenRoots.add(n.planet);
      return true;
    });

    // Build tree data recursively
    function buildNode(planet: string, visited: Set<string>): HierarchyDatum | null {
      if (visited.has(planet)) return null; // cycle guard
      visited.add(planet);

      const node = nodeMap.get(planet);
      if (!node) return null;

      const children: HierarchyDatum[] = [];
      for (const m of data.nodes) {
        if (m.dispositor === planet && m.planet !== planet) {
          const child = buildNode(m.planet, new Set(visited));
          if (child) children.push(child);
        }
      }

      return {
        name: node.planet,
        sign: node.sign,
        isFinal: node.is_final,
        inLoop: node.in_loop,
        mutualReception: mrSet.has(node.planet),
        children: children.length ? children : undefined,
      };
    }

    // If multiple roots, create a virtual root to hold them all
    let rootDatum: HierarchyDatum;
    if (uniqueRoots.length === 1) {
      const r = buildNode(uniqueRoots[0].planet, new Set());
      if (!r) return;
      rootDatum = r;
    } else {
      rootDatum = {
        name: '__root__',
        sign: '',
        isFinal: false,
        inLoop: false,
        mutualReception: false,
        children: uniqueRoots
          .map((r) => buildNode(r.planet, new Set()))
          .filter((n): n is HierarchyDatum => n !== null),
      };
    }

    // Create D3 hierarchy
    const root = d3.hierarchy<HierarchyDatum>(rootDatum);
    const treeLayout = d3.tree<HierarchyDatum>().nodeSize([60, 120]);
    treeLayout(root);

    // Compute bounding box
    let x0 = Infinity;
    let x1 = -Infinity;
    let y0 = Infinity;
    let y1 = -Infinity;
    root.each((d) => {
      const x = d.y ?? 0; // horizontal layout: y -> x-axis
      const y = d.x ?? 0; // x -> y-axis
      if (x < x0) x0 = x;
      if (x > x1) x1 = x;
      if (y < y0) y0 = y;
      if (y > y1) y1 = y;
    });

    const margin = 40;
    const width = y1 - y0 + margin * 2;
    const height = x1 - x0 + margin * 2;

    const g = svg
      .attr('viewBox', `${-margin} ${x0 - margin} ${width} ${height}`)
      .attr('style', 'width: 100%; height: 100%; min-height: 400px;')
      .append('g');

    // Links — horizontal: (y, x) mapping
    const linkGenerator = d3
      .linkHorizontal<d3.HierarchyLink<HierarchyDatum>, d3.HierarchyPointNode<HierarchyDatum>>()
      .x((d) => d.y ?? 0)
      .y((d) => d.x ?? 0);

    g.selectAll('.link')
      .data(root.links())
      .join('path')
      .attr('class', 'link')
      .attr('d', linkGenerator as any) // d3 link generator type mismatch
      .attr('fill', 'none')
      .attr('stroke', '#555')
      .attr('stroke-width', 1.5);

    // Node groups
    const nodeGroup = g
      .selectAll('.node')
      .data(root.descendants())
      .join('g')
      .attr('class', 'node')
      .attr('transform', (d) => `translate(${d.y},${d.x})`);

    // Circles
    nodeGroup
      .append('circle')
      .attr('r', 20)
      .attr('fill', (d: d3.HierarchyNode<HierarchyDatum>) => {
        const data = (d as d3.HierarchyPointNode<HierarchyDatum>).data;
        if (data.isFinal) return '#f0c040'; // gold for final dispositors
        if (data.mutualReception) return '#40a0f0'; // blue for mutual reception
        return '#333';
      })
      .attr('stroke', '#e0e0e0')
      .attr('stroke-width', 1.5);

    // Planet glyph text
    nodeGroup
      .append('text')
      .attr('text-anchor', 'middle')
      .attr('dominant-baseline', 'central')
      .attr('fill', '#e0e0e0')
      .attr('font-size', '18px')
      .attr('pointer-events', 'none')
      .text((d) => (d.data.name === '__root__' ? '' : planetGlyph(d.data.name)));

    // Sign label below node
    nodeGroup
      .filter((d) => d.data.name !== '__root__')
      .append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', 32)
      .attr('fill', '#999')
      .attr('font-size', '11px')
      .text((d) => d.data.sign);

    // Tooltips
    nodeGroup
      .filter((d) => d.data.name !== '__root__')
      .append('title')
      .text((d) => {
        const parts = [`${d.data.name} in ${d.data.sign}`];
        if (d.data.isFinal) parts.push('Final dispositor');
        if (d.data.mutualReception) parts.push('Mutual reception');
        if (d.data.inLoop) parts.push('In loop');
        parts.push(`Dispositor: ${d.data.sign} → ${nodeMap.get(d.data.name)?.dispositor ?? '?'}`);
        return parts.join('\n');
      });
  }, [data]);

  return <svg ref={svgRef} style={{ width: '100%', height: '100%', minHeight: 400 }} />;
}