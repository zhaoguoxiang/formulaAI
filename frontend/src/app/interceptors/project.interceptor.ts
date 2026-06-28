import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { ProjectStateService } from '../services/project-state.service';

/**
 * HTTP interceptor that attaches the X-Project-Id header to all API requests
 * except those targeting the global /api/projects endpoint.
 */
export const projectInterceptor: HttpInterceptorFn = (req, next) => {
  // Skip for project management endpoints (global scope)
  if (req.url.includes('/api/projects')) {
    return next(req);
  }

  // Skip for non-API requests
  if (!req.url.includes('/api/')) {
    return next(req);
  }

  const projectState = inject(ProjectStateService);
  const projectId = projectState.currentProjectId();

  if (!projectId) {
    return next(req);
  }

  const modifiedReq = req.clone({
    setHeaders: {
      'X-Project-Id': projectId,
    },
  });

  return next(modifiedReq);
};
