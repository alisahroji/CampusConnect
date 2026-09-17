import { useNavigate } from 'react-router-dom';

const ProjectCard = ({ project }) => {
  const navigate = useNavigate();

  // Tech stack disimpan sebagai string dipisah koma (mis. "React,Go,PostgreSQL")
  const techStack = (project.tech_stack || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean);

  return (
    <article
      onClick={() => navigate(`/projects/${project.id}`)}
      className="flex flex-col bg-white rounded-2xl border border-[#E2E8F0] shadow-sm hover:shadow-lg hover:-translate-y-1 transition-all duration-300 overflow-hidden cursor-pointer"
    >
      {/* Header / gambar */}
      <div className="h-40 bg-gradient-to-br from-[#112320] to-[#1E293B] flex items-center justify-center text-[#F8F9FA]">
        {project.image_url ? (
          <img src={project.image_url} alt={project.title} className="w-full h-full object-cover" />
        ) : (
          <span className="font-display text-4xl font-semibold text-[#D97757]">
            {project.title?.charAt(0)?.toUpperCase() ?? 'P'}
          </span>
        )}
      </div>

      {/* Isi kartu */}
      <div className="p-6 flex flex-col flex-1">
        <div className="flex items-start justify-between gap-2 mb-2">
          <h3 className="font-display text-xl font-semibold text-[#1E293B] leading-snug">
            {project.title}
          </h3>
          <span className="shrink-0 text-[10px] font-bold uppercase tracking-wider text-[#94A3B8] border border-[#E2E8F0] rounded-full px-2 py-0.5">
            {project.status || 'published'}
          </span>
        </div>

        <p className="text-sm text-[#64748B] leading-relaxed line-clamp-3 mb-4">
          {project.description}
        </p>

        {/* Pills tech stack */}
        {techStack.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-4">
            {techStack.map((tag) => (
              <span
                key={tag}
                className="text-xs font-semibold text-[#112320] bg-[#D97757]/10 border border-[#D97757]/20 rounded-full px-3 py-1"
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* Link demo & repo */}
        <div className="mt-auto flex items-center gap-4 pt-2">
          {project.demo_url && (
            <a
              href={project.demo_url}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="text-xs font-bold text-[#D97757] hover:text-[#C26244] transition-colors"
            >
              Demo ↗
            </a>
          )}
          {project.repo_url && (
            <a
              href={project.repo_url}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="text-xs font-bold text-[#64748B] hover:text-[#1E293B] transition-colors"
            >
              Repo ↗
            </a>
          )}
        </div>
      </div>
    </article>
  );
};

export default ProjectCard;
